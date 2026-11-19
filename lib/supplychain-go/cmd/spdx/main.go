package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/spdxnamespace"
	spdxJson "github.com/spdx/tools-golang/json"
	spdxTV "github.com/spdx/tools-golang/tagvalue"
	spdxYaml "github.com/spdx/tools-golang/yaml"
)

func main() {
	var outPath, graphPath, classificationsPath, format, buildStatusPath, documentNamespacePath, validatorPath string
	flag.StringVar(&outPath, "out", "", "The path to write the generated SPDX SBOM.")
	flag.StringVar(&graphPath, "graph", "", "The path to the graph JSON file.")
	flag.StringVar(&classificationsPath, "classifications", "", "The path to the classifications JSON file.")
	flag.StringVar(&format, "format", "json", "The output format of the SPDX SBOM: json, yaml and tag-value write SPDX 2.3, json-ld writes SPDX 3.0.1.")
	flag.StringVar(&buildStatusPath, "created_from_status_file", "", "Path to a Bazel volatile-status.txt file to read a BUILD_TIMESTAMP from for creationInfo.created. If unset, or the file has no BUILD_TIMESTAMP key, a fixed deterministic timestamp (the Unix epoch) is used instead, matching Bazel's own --nostamp default.")
	flag.StringVar(&documentNamespacePath, "document_namespace_file", "", "Optional path to a file holding the document namespace (an absolute URI without a fragment). It replaces the namespace derived from the subject's purl, and should be unique to each generated document.")
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

	documentNamespace, err := spdxnamespace.ReadFile(documentNamespacePath)
	if err != nil {
		panic(fmt.Errorf("reading document namespace: %w", err))
	}

	var buf bytes.Buffer
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	switch format {
	case "json", "yaml", "tag-value":
		doc, err := GenerateDocument23(graph, classifications, created)
		must(err)
		if documentNamespace != "" {
			doc.DocumentNamespace = documentNamespace
		}
		switch format {
		case "json":
			must(spdxJson.Write(doc, &buf))
		case "yaml":
			must(spdxYaml.Write(doc, &buf))
		case "tag-value":
			must(spdxTV.Write(doc, &buf))
		}
	case "json-ld":
		doc, err := GenerateDocument301(graph, classifications, created)
		must(err)
		if documentNamespace != "" {
			doc.ID = documentNamespace
		}
		must(writeDocument(doc, &buf))
	default:
		panic(fmt.Sprintf("'%s' is not a supported format", format))
	}

	must(os.WriteFile(outPath, buf.Bytes(), 0664))

	if validatorPath != "" {
		must(sbom.RunValidator(validatorPath, outPath))
	}
}

// readBuildTimestamp reads the BUILD_TIMESTAMP of a Bazel volatile-status.txt
// file for creationInfo.created; see sbom.ReadBuildTimestamp.
func readBuildTimestamp(path string) (time.Time, error) {
	return sbom.ReadBuildTimestamp(path)
}
