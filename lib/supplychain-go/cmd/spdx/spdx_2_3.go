package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	spdx2 "github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
)

// GenerateDocument23 builds an SPDX 2.3 document from the graph.
func GenerateDocument23(graph sbom.GraphConfig, classifications sbom.Classifications, created time.Time) (*spdx2.Document, error) {
	graph = sbom.NormalizeGraph(graph)
	spdxPackages := make([]*spdx2.Package, 0)
	labelToID := make(map[string]string)
	purlToID := make(map[string]string)
	purlToPackage := make(map[string]*spdx2.Package)
	licenseOf := make(map[*spdx2.Package]*supplychain.LicenseAttribute)
	patchesOf := make(map[*spdx2.Package][]*supplychain.PatchAttribute)
	idx := 0

	// Identify the document as a whole (name/namespace, below) by its
	// subject: preferably the root component; failing that (e.g.
	// require_root_metadata = False and the root has no metadata), the
	// first package created.
	var subjectName, subjectPURL string

	// Helper function to create package from node
	createPackage := func(node *sbom.NodeConfig) (*spdx2.Package, error) {
		if node.MetadataFile == "" {
			return nil, nil
		}

		pkgMetadata, err := supplychain.ReadPackageMetadataFromFile(node.MetadataFile)
		if err != nil {
			return nil, err
		}

		purl := pkgMetadata.GetPURL()
		if subjectPURL == "" {
			subjectName = purl.Name
			subjectPURL = purl.String()
		}
		licenseAttribute, err := supplychain.GetLicenseAttribute(pkgMetadata)
		if err != nil {
			return nil, fmt.Errorf("error reading license of %s: %w", node.Label, err)
		}
		if elementID, exists := purlToID[purl.String()]; exists {
			labelToID[node.Label] = elementID
			// Keep one package per PURL, but do not lose a license that only
			// a later metadata file declares.
			if existing := purlToPackage[purl.String()]; existing != nil && licenseAttribute != nil && licenseOf[existing] == nil {
				licenseOf[existing] = licenseAttribute
			}
			return nil, nil
		}

		elementID := fmt.Sprintf("dep-%d", idx)
		idx++

		downloadLocation := "NOASSERTION"
		qualifiers := purl.Qualifiers.Map()
		if qualifier := strings.TrimSpace(qualifiers["download_url"]); qualifier != "" {
			downloadLocation = qualifier
		} else if qualifier := strings.TrimSpace(qualifiers["url_download"]); qualifier != "" {
			downloadLocation = qualifier
		}

		pkg := &spdx2.Package{
			PackageSPDXIdentifier:   common.ElementID(elementID),
			PackageDownloadLocation: downloadLocation,
			PackageLicenseConcluded: "NOASSERTION",
			PackageLicenseDeclared:  "NOASSERTION",
			PackageExternalReferences: []*spdx2.PackageExternalReference{
				{
					Category: "PACKAGE-MANAGER",
					RefType:  "purl",
					Locator:  purl.String(),
				},
			},
			PackageName:    purl.Name,
			PackageVersion: purl.Version,
		}

		cpeAttribute, err := supplychain.GetCPEAttribute(pkgMetadata)
		if err != nil {
			return nil, fmt.Errorf("error reading CPE of %s: %w", node.Label, err)
		}
		if cpeAttribute != nil {
			for _, cpe := range *cpeAttribute {
				version, err := supplychain.CPEVersionOf(cpe)
				if err != nil {
					return nil, fmt.Errorf("error reading CPE of %s: %w", node.Label, err)
				}
				refType := "cpe23Type"
				if version == supplychain.CPE22 {
					refType = "cpe22Type"
				}
				pkg.PackageExternalReferences = append(pkg.PackageExternalReferences, &spdx2.PackageExternalReference{
					Category: "SECURITY",
					RefType:  refType,
					Locator:  cpe,
				})
			}
		}

		if licenseAttribute != nil {
			licenseOf[pkg] = licenseAttribute
		}

		integrity, err := supplychain.GetIntegrityAttribute(pkgMetadata)
		if err != nil {
			return nil, fmt.Errorf("error reading integrity of %s: %w", node.Label, err)
		}
		if integrity != nil {
			checksums, err := integrity.Checksums()
			if err != nil {
				return nil, fmt.Errorf("error reading integrity of %s: %w", node.Label, err)
			}
			for _, checksum := range checksums {
				pkg.PackageChecksums = append(pkg.PackageChecksums, common.Checksum{
					Algorithm: common.ChecksumAlgorithm(strings.ToUpper(checksum.Algorithm)),
					Value:     checksum.Hex,
				})
			}
		}

		patches, err := supplychain.GetPatchAttributes(pkgMetadata)
		if err != nil {
			return nil, fmt.Errorf("error reading patches of %s: %w", node.Label, err)
		}
		patchesOf[pkg] = patches

		componentType, err := supplychain.GetComponentTypeAttribute(pkgMetadata)
		if err != nil {
			return nil, fmt.Errorf("error reading component type of %s: %w", node.Label, err)
		}
		// SPDX 2.3 purposes are the upper-case component types.
		pkg.PrimaryPackagePurpose = strings.ToUpper(componentType)

		purlToPackage[purl.String()] = pkg
		labelToID[node.Label] = elementID
		purlToID[purl.String()] = elementID
		return pkg, nil
	}

	// Add root component
	if classifications.RootComponent != nil {
		pkg, err := createPackage(classifications.RootComponent)
		if err != nil {
			return nil, err
		}
		if pkg != nil {
			spdxPackages = append(spdxPackages, pkg)
		}
	}

	// Add all dependencies (direct + transitive)
	for i := range classifications.Dependencies.Direct {
		pkg, err := createPackage(&classifications.Dependencies.Direct[i])
		if err != nil {
			return nil, err
		}
		if pkg != nil {
			spdxPackages = append(spdxPackages, pkg)
		}
	}
	for i := range classifications.Dependencies.Transitive {
		pkg, err := createPackage(&classifications.Dependencies.Transitive[i])
		if err != nil {
			return nil, err
		}
		if pkg != nil {
			spdxPackages = append(spdxPackages, pkg)
		}
	}

	// Build Relationships from graph edges
	relationships := make([]*spdx2.Relationship, 0, len(graph.Edges))
	seenRelationship := make(map[[3]string]bool)
	for _, edge := range graph.Edges {
		fromID, fromOk := labelToID[edge.From]
		toID, toOk := labelToID[edge.To]

		if !fromOk || !toOk || fromID == toID {
			// A self-reference can appear if two distinct graph nodes
			// resolved to the same package (e.g. duplicate metadata).
			continue
		}
		key := [3]string{fromID, toID, edge.Involvement}
		if seenRelationship[key] {
			continue
		}
		seenRelationship[key] = true
		refA, refB, relationship := fromID, toID, "DEPENDS_ON"
		switch edge.Involvement {
		case "build":
			// SPDX 2.3's BUILD_TOOL_OF is directed from the tool to its
			// consumer, the reverse of the graph dependency edge.
			refA, refB, relationship = toID, fromID, "BUILD_TOOL_OF"
		case "embedded":
			relationship = "CONTAINS"
		}
		relationships = append(relationships, &spdx2.Relationship{
			RefA:         common.MakeDocElementID("", refA),
			RefB:         common.MakeDocElementID("", refB),
			Relationship: relationship,
		})
	}

	// Add DESCRIBES relationship from DOCUMENT to root component
	if classifications.RootComponent != nil {
		if rootID, ok := labelToID[classifications.RootComponent.Label]; ok {
			relationships = append(relationships, &spdx2.Relationship{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("", rootID),
				Relationship: "DESCRIBES",
			})
		}
	}

	otherLicenses, err := declaredLicenses23(spdxPackages, licenseOf)
	if err != nil {
		return nil, err
	}

	patchPackages, patchRelationships, err := patchPackages23(spdxPackages, patchesOf)
	if err != nil {
		return nil, err
	}
	spdxPackages = append(spdxPackages, patchPackages...)
	relationships = append(relationships, patchRelationships...)

	if subjectName == "" {
		subjectName = "sbom"
		subjectPURL = "sbom"
	}

	doc := spdx2.Document{
		SPDXIdentifier: "DOCUMENT",
		SPDXVersion:    "SPDX-2.3",
		// 6.2: mandatory, and specifically "CC0-1.0" per the spec. This is
		// the license of the SPDX *metadata* document itself, independent of
		// the licenses of the packages it describes.
		DataLicense:  "CC0-1.0",
		DocumentName: subjectName,
		// 6.5: mandatory, must be unique to this document. Derived
		// deterministically from the subject's purl (rather than e.g. a
		// random UUID) so that re-generating the SBOM for an unchanged
		// dependency graph produces byte-identical output and stays
		// Bazel-cacheable.
		DocumentNamespace: "https://spdx.org/spdxdocs/" + url.PathEscape(subjectPURL),
		Packages:          spdxPackages,
		OtherLicenses:     otherLicenses,
		Relationships:     relationships,
		CreationInfo: &spdx2.CreationInfo{
			Creators: []common.Creator{
				{Creator: "Bazel Supply Chain Tools SPDX generator", CreatorType: "Tool"},
			},
			Created: created.Format(time.RFC3339),
		},
	}

	return &doc, nil
}
