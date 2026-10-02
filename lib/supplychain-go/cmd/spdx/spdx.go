package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	spdxJson "github.com/spdx/tools-golang/json"
	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
	spdxTV "github.com/spdx/tools-golang/tagvalue"
	spdxYaml "github.com/spdx/tools-golang/yaml"
)

func main() {
	var outPath, graphPath, classificationsPath, format, buildStatusPath string
	flag.StringVar(&outPath, "out", "", "The path to write the generated SPDX SBOM.")
	flag.StringVar(&graphPath, "graph", "", "The path to the graph JSON file.")
	flag.StringVar(&classificationsPath, "classifications", "", "The path to the classifications JSON file.")
	flag.StringVar(&format, "format", "json", "The output format of the SPDX SBOM.")
	flag.StringVar(&buildStatusPath, "created_from_status_file", "", "Path to a Bazel volatile-status.txt file to read a BUILD_TIMESTAMP from for creationInfo.created. If unset, or the file has no BUILD_TIMESTAMP key, a fixed deterministic timestamp (the Unix epoch) is used instead, matching Bazel's own --nostamp default.")
	flag.Parse()

	if graphPath == "" || classificationsPath == "" {
		panic("both --graph and --classifications flags are required")
	}

	graphBytes, err := os.ReadFile(graphPath)
	if err != nil {
		panic(fmt.Errorf("reading graph: %w", err))
	}

	var graph sbom.GraphConfig
	if err := json.Unmarshal(graphBytes, &graph); err != nil {
		panic(fmt.Errorf("parsing graph: %w", err))
	}

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

	created, err := readBuildTimestamp(buildStatusPath)
	if err != nil {
		panic(fmt.Errorf("reading build status file: %w", err))
	}

	doc, err := GenerateDocument(graph, classifications, created)
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

func GenerateDocument(graph sbom.GraphConfig, classifications sbom.Classifications, created time.Time) (*spdx.Document, error) {
	spdxPackages := make([]*spdx.Package, 0)
	labelToID := make(map[string]string)
	idx := 0

	// Identify the document as a whole (name/namespace, below) by its
	// subject: preferably the root component; failing that (e.g.
	// require_root_metadata = False and the root has no metadata), the
	// first package created.
	var subjectName, subjectPURL string

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

		purl := pkgMetadata.GetPURL()
		if subjectPURL == "" {
			subjectName = purl.Name
			subjectPURL = purl.String()
		}

		pkg := &spdx.Package{
			PackageSPDXIdentifier: common.ElementID(elementID),
			PackageExternalReferences: []*spdx.PackageExternalReference{
				{
					Category: "PACKAGE-MANAGER",
					RefType:  "purl",
					Locator:  purl.String(),
				},
			},
			PackageName: purl.Name,
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
		relationships = append(relationships, &spdx.Relationship{
			RefA:         common.MakeDocElementID("", fromID),
			RefB:         common.MakeDocElementID("", toID),
			Relationship: "DEPENDS_ON",
		})
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

	if subjectName == "" {
		subjectName = "sbom"
		subjectPURL = "sbom"
	}

	doc := spdx.Document{
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
		CreationInfo: &spdx.CreationInfo{
			Creators: []common.Creator{
				{Creator: "Bazel Supply Chain Tools SPDX generator", CreatorType: "Tool"},
			},
			Created: created.Format(time.RFC3339),
		},
	}

	return &doc, nil
}

// readBuildTimestamp reads a BUILD_TIMESTAMP key out of a Bazel
// volatile-status.txt file (see ctx.version_file / --workspace_status_command),
// for use as creationInfo.created. If path is empty, or the file has no
// BUILD_TIMESTAMP key, it returns the Unix epoch: the same fixed value Bazel
// itself substitutes for BUILD_TIMESTAMP when building without --stamp, so
// SBOM generation stays deterministic and cacheable unless the caller
// explicitly opts into a real timestamp via --stamp.
func readBuildTimestamp(path string) (time.Time, error) {
	if path == "" {
		return time.Unix(0, 0).UTC(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		key, value, found := strings.Cut(line, " ")
		if !found || key != "BUILD_TIMESTAMP" {
			continue
		}
		seconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("parsing BUILD_TIMESTAMP %q: %w", value, err)
		}
		return time.Unix(seconds, 0).UTC(), nil
	}

	return time.Unix(0, 0).UTC(), nil
}
