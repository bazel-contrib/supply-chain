package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/coverage"
)

type resolutionFile struct {
	Packages []resolution `json:"packages"`
}

type resolution struct {
	PURL              string `json:"purl"`
	LicenseExpression string `json:"license_expression"`
}

func main() {
	var graphPath, resolutionsPath, outputPath string
	flag.StringVar(&graphPath, "graph", "", "Graph JSON from gather_metadata")
	flag.StringVar(&resolutionsPath, "resolutions", "", "Optional license resolver JSON")
	flag.StringVar(&outputPath, "output", "", "Coverage report JSON")
	flag.Parse()
	if graphPath == "" || outputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --graph and --output are required")
		os.Exit(2)
	}

	graphBytes, err := os.ReadFile(graphPath)
	if err != nil {
		fail(err)
	}
	var graph coverage.GraphConfig
	if err := json.Unmarshal(graphBytes, &graph); err != nil {
		fail(fmt.Errorf("parse graph: %w", err))
	}

	resolved := make(map[string][]string)
	if resolutionsPath != "" {
		resolutionBytes, err := os.ReadFile(resolutionsPath)
		if err != nil {
			fail(err)
		}
		var resolutions resolutionFile
		if err := json.Unmarshal(resolutionBytes, &resolutions); err != nil {
			fail(fmt.Errorf("parse resolutions: %w", err))
		}
		for _, result := range resolutions.Packages {
			if result.PURL != "" && result.LicenseExpression != "" {
				resolved[result.PURL] = []string{result.LicenseExpression}
			}
		}
	}

	report, err := coverage.BuildReport(graph, resolved)
	if err != nil {
		fail(err)
	}
	outputBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fail(err)
	}
	outputBytes = append(outputBytes, '\n')
	if err := os.WriteFile(outputPath, outputBytes, 0o644); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
