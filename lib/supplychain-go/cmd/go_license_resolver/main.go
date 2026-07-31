// Command go_license_resolver resolves Go package licenses from a Bazel
// metadata graph and the fetched repositories in the Bazel output base.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/coverage"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/go_bazel"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/license_resolver"
)

func main() {
	var graphPath, outputPath, outputBase, workspace string
	flag.StringVar(&graphPath, "graph", "", "Graph JSON from gather_metadata")
	flag.StringVar(&outputPath, "output", "", "Resolver JSON output")
	flag.StringVar(&outputBase, "output-base", "", "Bazel output base; defaults to BAZEL_OUTPUT_BASE or bazel info output_base")
	flag.StringVar(&workspace, "workspace", ".", "Workspace in which to run bazel info when output-base is omitted")
	flag.Parse()
	if graphPath == "" || outputPath == "" {
		fail(fmt.Errorf("--graph and --output are required"))
	}

	graphBytes, err := os.ReadFile(graphPath)
	if err != nil {
		fail(fmt.Errorf("read graph: %w", err))
	}
	var graph coverage.GraphConfig
	if err := json.Unmarshal(graphBytes, &graph); err != nil {
		fail(fmt.Errorf("parse graph: %w", err))
	}

	if outputBase == "" {
		outputBase = os.Getenv("BAZEL_OUTPUT_BASE")
	}
	if outputBase == "" {
		outputBase, err = discoverOutputBase(workspace)
		if err != nil {
			fail(err)
		}
	}

	packages, err := go_bazel.BuildRequest(graph, outputBase)
	if err != nil {
		fail(err)
	}
	classifier, err := license_resolver.NewClassifier(license_resolver.Options{})
	if err != nil {
		fail(err)
	}
	response, err := classifier.Resolve(packages)
	if err != nil {
		fail(err)
	}

	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(outputPath, append(output, '\n'), 0o644); err != nil {
		fail(fmt.Errorf("write output: %w", err))
	}
}

func discoverOutputBase(workspace string) (string, error) {
	bazel := os.Getenv("BAZEL")
	if bazel == "" {
		bazel = "bazel"
	}
	command := exec.Command(bazel, "info", "output_base")
	command.Dir = filepath.Clean(workspace)
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("discover Bazel output base: %w", err)
	}
	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("bazel info output_base returned an empty path")
	}
	return result, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
