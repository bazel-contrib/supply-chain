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

		elementID := fmt.Sprintf("dep-%d", idx)
		idx++

		purl := pkgMetadata.GetPURL()
		if subjectPURL == "" {
			subjectName = purl.Name
			subjectPURL = purl.String()
		}

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

		labelToID[node.Label] = elementID
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
	seenRelationship := make(map[string]map[string]bool)
	for _, edge := range graph.Edges {
		fromID, fromOk := labelToID[edge.From]
		toID, toOk := labelToID[edge.To]

		if !fromOk || !toOk || fromID == toID {
			// A self-reference can appear if two distinct graph nodes
			// resolved to the same package (e.g. duplicate metadata).
			continue
		}
		if seenRelationship[fromID] == nil {
			seenRelationship[fromID] = make(map[string]bool)
		}
		if seenRelationship[fromID][toID] {
			continue
		}
		seenRelationship[fromID][toID] = true
		relationships = append(relationships, &spdx2.Relationship{
			RefA:         common.MakeDocElementID("", fromID),
			RefB:         common.MakeDocElementID("", toID),
			Relationship: "DEPENDS_ON",
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
