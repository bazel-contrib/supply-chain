package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	spdxJson "github.com/spdx/tools-golang/json"
	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
	spdxTV "github.com/spdx/tools-golang/tagvalue"
	spdxYaml "github.com/spdx/tools-golang/yaml"
)

func main() {
	var outPath, graphPath, classificationsPath, format string
	flag.StringVar(&outPath, "out", "", "The path to write the generated SPDX SBOM.")
	flag.StringVar(&graphPath, "graph", "", "The path to the graph JSON file.")
	flag.StringVar(&classificationsPath, "classifications", "", "The path to the classifications JSON file.")
	flag.StringVar(&format, "format", "json", "The output format of the SPDX SBOM.")
	flag.Parse()

	if graphPath == "" || classificationsPath == "" {
		panic("both --graph and --classifications flags are required")
	}

	// Read graph
	graphBytes, err := os.ReadFile(graphPath)
	if err != nil {
		panic(fmt.Errorf("reading graph: %w", err))
	}

	var graph sbom.GraphConfig
	if err := json.Unmarshal(graphBytes, &graph); err != nil {
		panic(fmt.Errorf("parsing graph: %w", err))
	}

	// Read classifications
	classBytes, err := os.ReadFile(classificationsPath)
	if err != nil {
		panic(fmt.Errorf("reading classifications: %w", err))
	}

	var classifications sbom.Classifications
	if err := json.Unmarshal(classBytes, &classifications); err != nil {
		panic(fmt.Errorf("parsing classifications: %w", err))
	}

	out, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	doc, err := GenerateDocument(graph, classifications)
	if err != nil {
		panic(err)
	}

	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	switch format {
	case "json":
		must(spdxJson.Write(doc, out))
	case "yaml":
		must(spdxYaml.Write(doc, out))
	case "tag-value":
		must(spdxTV.Write(doc, out))
	default:
		panic(fmt.Sprintf("'%s' is not a supported format", format))
	}
}

func GenerateDocument(graph sbom.GraphConfig, classifications sbom.Classifications) (*spdx.Document, error) {
	spdxPackages := make([]*spdx.Package, 0)
	labelToID := make(map[string]string)
	idx := 0

	// Helper function to create package from node
	createPackage := func(node *sbom.NodeConfig) (*spdx.Package, error) {
		if node.MetadataFile == "" {
			return nil, nil
		}

		pkgMetadata, err := supplychain.ReadPackageMetadataFromFile(node.MetadataFile)
		if err != nil {
			return nil, err
		}

		elementID := fmt.Sprintf("dep-%d", idx)
		idx++

		pkg := &spdx.Package{
			PackageSPDXIdentifier: common.ElementID(elementID),
			PackageExternalReferences: []*spdx.PackageExternalReference{
				{
					Category: "PACKAGE-MANAGER",
					RefType:  "purl",
					Locator:  pkgMetadata.GetPURL().String(),
				},
			},
			PackageName: pkgMetadata.GetPURL().Name,
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
	relationships := make([]*spdx.Relationship, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		fromID, fromOk := labelToID[edge.From]
		toID, toOk := labelToID[edge.To]

		if fromOk && toOk {
			relationships = append(relationships, &spdx.Relationship{
				RefA:         common.MakeDocElementID("", fromID),
				RefB:         common.MakeDocElementID("", toID),
				Relationship: "DEPENDS_ON",
			})
		}
	}

	// Add DESCRIBES relationship from DOCUMENT to root component
	if classifications.RootComponent != nil {
		if rootID, ok := labelToID[classifications.RootComponent.Label]; ok {
			relationships = append(relationships, &spdx.Relationship{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("", rootID),
				Relationship: "DESCRIBES",
			})
		}
	}

	doc := spdx.Document{
		SPDXIdentifier: "DOCUMENT",
		SPDXVersion:    "SPDX-2.3",
		Packages:       spdxPackages,
		Relationships:  relationships,
		CreationInfo:   &spdx.CreationInfo{},
	}

	return &doc, nil
}
