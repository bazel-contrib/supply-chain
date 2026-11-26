package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	spdx "github.com/spdx/tools-golang/spdx/v3/v3_0"
)

// GenerateDocument301 builds an SPDX 3.0.1 document from the graph.
func GenerateDocument301(graph sbom.GraphConfig, classifications sbom.Classifications, created time.Time) (*spdx.Document, error) {
	graph = sbom.NormalizeGraph(graph)
	labelToPackage := make(map[string]*spdx.Package)
	purlToPackage := make(map[string]*spdx.Package)
	var packages []*spdx.Package
	licenseOf := make(map[*spdx.Package]*supplychain.LicenseAttribute)
	patchesOf := make(map[*spdx.Package][]*supplychain.PatchAttribute)

	// Identify the document as a whole (name/namespace, below) by its
	// subject: preferably the root component; failing that (e.g.
	// require_root_metadata = False and the root has no metadata), the
	// first package created.
	var subjectName, subjectPURL string

	createPackage := func(node *sbom.NodeConfig) error {
		if node.MetadataFile == "" {
			return nil
		}

		pkgMetadata, err := supplychain.ReadPackageMetadataFromFile(node.MetadataFile)
		if err != nil {
			return err
		}

		purl := pkgMetadata.GetPURL()
		if subjectPURL == "" {
			subjectName = purl.Name
			subjectPURL = purl.String()
		}
		licenseAttribute, err := supplychain.GetLicenseAttribute(pkgMetadata)
		if err != nil {
			return fmt.Errorf("error reading license of %s: %w", node.Label, err)
		}
		if existing, exists := purlToPackage[purl.String()]; exists {
			labelToPackage[node.Label] = existing
			// Keep one package per PURL, but do not lose a license that only
			// a later metadata file declares.
			if licenseAttribute != nil && licenseOf[existing] == nil {
				licenseOf[existing] = licenseAttribute
			}
			return nil
		}

		qualifiers := purl.Qualifiers.Map()
		downloadLocation := strings.TrimSpace(qualifiers["download_url"])
		if downloadLocation == "" {
			downloadLocation = strings.TrimSpace(qualifiers["url_download"])
		}

		pkg := &spdx.Package{
			ID:               fmt.Sprintf("dep-%d", len(packages)),
			Name:             purl.Name,
			Version:          purl.Version,
			PackageURL:       spdx.URI(purl.String()),
			DownloadLocation: spdx.URI(downloadLocation),
		}
		pkg.OriginatedBy = originatedBy(purl.Type, purl.Namespace)
		integrity, err := supplychain.GetIntegrityAttribute(pkgMetadata)
		if err != nil {
			return fmt.Errorf("error reading integrity of %s: %w", node.Label, err)
		}
		if integrity != nil {
			checksums, err := integrity.Checksums()
			if err != nil {
				return fmt.Errorf("error reading integrity of %s: %w", node.Label, err)
			}
			for _, checksum := range checksums {
				pkg.VerifiedUsing = append(pkg.VerifiedUsing, &spdx.Hash{
					Algorithm: hashAlgorithms[checksum.Algorithm],
					Value:     checksum.Hex,
				})
			}
		}
		cpeAttribute, err := supplychain.GetCPEAttribute(pkgMetadata)
		if err != nil {
			return fmt.Errorf("error reading CPE of %s: %w", node.Label, err)
		}
		if cpeAttribute != nil {
			for _, cpe := range *cpeAttribute {
				version, err := supplychain.CPEVersionOf(cpe)
				if err != nil {
					return fmt.Errorf("error reading CPE of %s: %w", node.Label, err)
				}
				identifierType := spdx.ExternalIdentifierType_Cpe23
				if version == supplychain.CPE22 {
					identifierType = spdx.ExternalIdentifierType_Cpe22
				}
				pkg.ExternalIdentifiers = append(pkg.ExternalIdentifiers, &spdx.ExternalIdentifier{
					Type:       identifierType,
					Identifier: cpe,
				})
			}
		}
		if licenseAttribute != nil {
			licenseOf[pkg] = licenseAttribute
		}
		patches, err := supplychain.GetPatchAttributes(pkgMetadata)
		if err != nil {
			return fmt.Errorf("error reading patches of %s: %w", node.Label, err)
		}
		patchesOf[pkg] = patches
		packages = append(packages, pkg)
		labelToPackage[node.Label] = pkg
		purlToPackage[purl.String()] = pkg
		return nil
	}

	if classifications.RootComponent != nil {
		if err := createPackage(classifications.RootComponent); err != nil {
			return nil, err
		}
	}
	for i := range classifications.Dependencies.Direct {
		if err := createPackage(&classifications.Dependencies.Direct[i]); err != nil {
			return nil, err
		}
	}
	for i := range classifications.Dependencies.Transitive {
		if err := createPackage(&classifications.Dependencies.Transitive[i]); err != nil {
			return nil, err
		}
	}

	var relationships []spdx.AnyRelationship
	seenRelationship := make(map[[4]string]bool)
	for _, edge := range graph.Edges {
		from, fromOk := labelToPackage[edge.From]
		to, toOk := labelToPackage[edge.To]

		// A self-reference can appear if two distinct graph nodes resolved
		// to the same package (e.g. duplicate metadata).
		if !fromOk || !toOk || from == to {
			continue
		}
		key := [4]string{from.ID, to.ID, edge.Involvement, edge.Relationship}
		if seenRelationship[key] {
			continue
		}
		seenRelationship[key] = true
		id := fmt.Sprintf("rel-%d", len(relationships))
		switch edge.Involvement {
		case "build":
			// SPDX 3 models a build tool as a tool used by the consumer.
			relationships = append(relationships, &spdx.LifecycleScopedRelationship{
				ID:    id,
				From:  from,
				To:    spdx.ElementList{to},
				Type:  spdx.RelationshipType_UsesTool,
				Scope: spdx.LifecycleScopeType_Build,
			})
		case "embedded":
			relationships = append(relationships, &spdx.Relationship{
				ID:   id,
				From: from,
				To:   spdx.ElementList{to},
				Type: spdx.RelationshipType_Contains,
			})
		default:
			switch edge.Relationship {
			case "contains":
				relationships = append(relationships, &spdx.Relationship{ID: id, From: from, To: spdx.ElementList{to}, Type: spdx.RelationshipType_Contains})
			case "static_link":
				relationships = append(relationships, &spdx.Relationship{ID: id, From: from, To: spdx.ElementList{to}, Type: spdx.RelationshipType_HasStaticLink})
			case "dynamic_link":
				relationships = append(relationships, &spdx.Relationship{ID: id, From: from, To: spdx.ElementList{to}, Type: spdx.RelationshipType_HasDynamicLink})
			case "test_dependency", "dev_dependency":
				scope := spdx.LifecycleScopeType_Test
				if edge.Relationship == "dev_dependency" {
					scope = spdx.LifecycleScopeType_Development
				}
				relationships = append(relationships, &spdx.LifecycleScopedRelationship{
					ID: id, From: from, To: spdx.ElementList{to},
					Type: spdx.RelationshipType_DependsOn, Scope: scope,
				})
			default:
				relationships = append(relationships, &spdx.Relationship{
					ID: id, From: from, To: spdx.ElementList{to},
					Type: spdx.RelationshipType_DependsOn,
				})
			}
		}
	}

	licenses, licenseRelationships, err := declaredLicenses(packages, licenseOf, len(relationships))
	if err != nil {
		return nil, err
	}
	for _, relationship := range licenseRelationships {
		relationships = append(relationships, relationship)
	}

	patches, patchRelationships, err := patchFiles(packages, patchesOf, len(relationships))
	if err != nil {
		return nil, err
	}
	for _, relationship := range patchRelationships {
		relationships = append(relationships, relationship)
	}

	if subjectName == "" {
		subjectName = "sbom"
		subjectPURL = "sbom"
	}

	doc := spdx.NewDocument(
		spdx.ProfileIdentifierType_Software,
		subjectName,
		&spdx.SoftwareAgent{Name: generatorName},
		generatorTool(toolPURL),
	)
	// The document ID is mandatory and must be unique to this document.
	// Derived deterministically from the subject's purl (rather than e.g. a
	// random UUID) so that re-generating the SBOM for an unchanged dependency
	// graph produces byte-identical output and stays Bazel-cacheable.
	doc.ID = "https://spdx.org/spdxdocs/" + url.PathEscape(subjectPURL)
	doc.Name = subjectName
	doc.CreationInfo.(*spdx.CreationInfo).Created = created
	doc.DataLicense = &spdx.ListedLicense{
		ID:   "https://spdx.org/licenses/CC0-1.0",
		Name: "CC0-1.0",
	}

	// The document describes a single SBOM, which holds all packages and
	// relationships. The root component is the root of the SBOM.
	sbomElement := &spdx.SBOM{
		ID:        "sbom",
		SbomTypes: []spdx.SbomType{spdx.SbomType_Build},
	}
	if classifications.RootComponent != nil {
		if root, ok := labelToPackage[classifications.RootComponent.Label]; ok {
			sbomElement.RootElements = spdx.ElementList{root}
		}
	}
	for _, pkg := range packages {
		sbomElement.Elements = append(sbomElement.Elements, pkg)
	}
	for _, license := range licenses {
		sbomElement.Elements = append(sbomElement.Elements, license)
	}
	for _, patch := range patches {
		sbomElement.Elements = append(sbomElement.Elements, patch)
	}
	for _, relationship := range relationships {
		sbomElement.Elements = append(sbomElement.Elements, relationship)
	}
	doc.RootElements = spdx.ElementList{sbomElement}

	return doc, nil
}

// writeDocument writes the document as canonical JSON-LD. tools-golang orders
// the graph nodes and the members of element collections by map iteration, so
// they are sorted here to keep the output byte-identical between runs.
func writeDocument(doc *spdx.Document, w io.Writer) error {
	var raw bytes.Buffer
	if err := doc.Write(&raw); err != nil {
		return err
	}
	var out struct {
		Context string           `json:"@context"`
		Graph   []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal(raw.Bytes(), &out); err != nil {
		return err
	}
	nodeID := func(node map[string]any) string {
		for _, key := range []string{"spdxId", "@id"} {
			if id, ok := node[key].(string); ok {
				return id
			}
		}
		return ""
	}
	sort.SliceStable(out.Graph, func(i, j int) bool {
		return nodeID(out.Graph[i]) < nodeID(out.Graph[j])
	})
	for _, node := range out.Graph {
		if elements, ok := node["element"].([]any); ok {
			sort.Slice(elements, func(i, j int) bool {
				return fmt.Sprint(elements[i]) < fmt.Sprint(elements[j])
			})
		}
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(out)
}

const generatorName = "Bazel Supply Chain Tools SPDX generator"

// toolPURL is the purl of the generator, set at link time from the module
// name and version. It stays empty for builds that are not a release.
var toolPURL = ""

// generatorTool describes the generator as the tool that created the
// document.
func generatorTool(purl string) *spdx.Tool {
	tool := &spdx.Tool{Name: generatorName}
	if purl != "" {
		tool.ExternalIdentifiers = spdx.ExternalIdentifierList{
			&spdx.ExternalIdentifier{
				Type:       spdx.ExternalIdentifierType_PackageURL,
				Identifier: purl,
			},
		}
	}
	return tool
}

var hashAlgorithms = map[string]spdx.HashAlgorithm{
	"sha256": spdx.HashAlgorithm_Sha256,
	"sha384": spdx.HashAlgorithm_Sha384,
	"sha512": spdx.HashAlgorithm_Sha512,
}

// originatedBy names the organization the package comes from, as a fallback for
// a package that declares no originator: an explicit originator attribute (see
// #274) takes precedence over it. Only purl types whose namespace is the owner
// of the repository tell it; for others (e.g. the maven group or the npm scope)
// the namespace is not necessarily one.
func originatedBy(purlType, namespace string) spdx.AgentList {
	switch purlType {
	case "github", "gitlab", "bitbucket":
		if namespace != "" {
			return spdx.AgentList{&spdx.Organization{Name: namespace}}
		}
	}
	return nil
}
