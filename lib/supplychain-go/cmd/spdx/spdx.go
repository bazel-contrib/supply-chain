package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	spdx "github.com/spdx/tools-golang/spdx/v3/v3_0"
)

func main() {
	var outPath, graphPath, classificationsPath, buildStatusPath, validatorPath string
	flag.StringVar(&outPath, "out", "", "The path to write the generated SPDX SBOM.")
	flag.StringVar(&graphPath, "graph", "", "The path to the graph JSON file.")
	flag.StringVar(&classificationsPath, "classifications", "", "The path to the classifications JSON file.")
	flag.StringVar(&buildStatusPath, "created_from_status_file", "", "Path to a Bazel volatile-status.txt file to read a BUILD_TIMESTAMP from for creationInfo.created. If unset, or the file has no BUILD_TIMESTAMP key, a fixed deterministic timestamp (the Unix epoch) is used instead, matching Bazel's own --nostamp default.")
	flag.StringVar(&validatorPath, "validator", "", "Optional runfiles path or executable path to the SPDX tools-java validator wrapper.")
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

	created, err := readBuildTimestamp(buildStatusPath)
	if err != nil {
		panic(fmt.Errorf("reading build status file: %w", err))
	}

	doc, err := GenerateDocument(graph, classifications, created)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	must(writeDocument(doc, &buf))

	must(os.WriteFile(outPath, buf.Bytes(), 0664))

	if validatorPath != "" {
		must(sbom.RunValidator(validatorPath, outPath))
	}
}

// GenerateDocument builds an SPDX 3.0.1 document from the graph.
func GenerateDocument(graph sbom.GraphConfig, classifications sbom.Classifications, created time.Time) (*spdx.Document, error) {
	graph = sbom.NormalizeGraph(graph)
	labelToPackage := make(map[string]*spdx.Package)
	var packages []*spdx.Package

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
		packages = append(packages, pkg)
		labelToPackage[node.Label] = pkg
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

	var relationships []*spdx.Relationship
	seenRelationship := make(map[[2]string]bool)
	for _, edge := range graph.Edges {
		from, fromOk := labelToPackage[edge.From]
		to, toOk := labelToPackage[edge.To]

		// A self-reference can appear if two distinct graph nodes resolved
		// to the same package (e.g. duplicate metadata).
		if !fromOk || !toOk || from == to {
			continue
		}
		key := [2]string{from.ID, to.ID}
		if seenRelationship[key] {
			continue
		}
		seenRelationship[key] = true
		relationships = append(relationships, &spdx.Relationship{
			ID:   fmt.Sprintf("rel-%d", len(relationships)),
			From: from,
			To:   spdx.ElementList{to},
			Type: spdx.RelationshipType_DependsOn,
		})
	}

	if subjectName == "" {
		subjectName = "sbom"
		subjectPURL = "sbom"
	}

	doc := spdx.NewDocument(
		spdx.ProfileIdentifierType_Software,
		subjectName,
		&spdx.SoftwareAgent{Name: "Bazel Supply Chain Tools SPDX generator"},
		nil,
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
	sbomElement := &spdx.SBOM{ID: "sbom"}
	if classifications.RootComponent != nil {
		if root, ok := labelToPackage[classifications.RootComponent.Label]; ok {
			sbomElement.RootElements = spdx.ElementList{root}
		}
	}
	for _, pkg := range packages {
		sbomElement.Elements = append(sbomElement.Elements, pkg)
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
