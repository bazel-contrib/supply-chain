package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	spdxJson "github.com/spdx/tools-golang/json"
	spdxTV "github.com/spdx/tools-golang/tagvalue"
	spdxYaml "github.com/spdx/tools-golang/yaml"
)

func main() {
	var outPath, graphPath, classificationsPath, format, buildStatusPath, validatorPath string
	flag.StringVar(&outPath, "out", "", "The path to write the generated SPDX SBOM.")
	flag.StringVar(&graphPath, "graph", "", "The path to the graph JSON file.")
	flag.StringVar(&classificationsPath, "classifications", "", "The path to the classifications JSON file.")
	flag.StringVar(&format, "format", "json", "The output format of the SPDX SBOM: json, yaml and tag-value write SPDX 2.3, json-ld writes SPDX 3.0.1.")
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
		must(writeDocument(doc, &buf))
	default:
		panic(fmt.Sprintf("'%s' is not a supported format", format))
	}

	must(os.WriteFile(outPath, buf.Bytes(), 0664))

	if validatorPath != "" {
		must(sbom.RunValidator(validatorPath, outPath))
	}
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
