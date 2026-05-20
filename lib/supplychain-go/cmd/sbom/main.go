package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

func main() {
	var inputPath, outputPath string
	var allowMissingRootMetadata bool
	flag.StringVar(&inputPath, "input", "", "Path to graph-only JSON from gather_metadata")
	flag.StringVar(&outputPath, "output", "", "Path to write SBOM classifications")
	flag.BoolVar(&allowMissingRootMetadata, "allow_missing_root_metadata", false, "Allow SBOM generation even if root target has no metadata")
	flag.Parse()

	if inputPath == "" || outputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --input and --output flags are required")
		flag.Usage()
		os.Exit(1)
	}

	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var graphConfig sbom.GraphConfig
	if err := json.Unmarshal(inputBytes, &graphConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing graph JSON: %v\n", err)
		os.Exit(1)
	}

	classifications, err := sbom.ComputeClassifications(graphConfig, allowMissingRootMetadata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	outputBytes, err := json.MarshalIndent(classifications, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding classifications: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, outputBytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}
}
