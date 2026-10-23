package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	resolver "github.com/bazel-contrib/supply-chain/lib/supplychain-go/license_resolver"
)

func main() {
	var inputPath, outputPath string
	flag.StringVar(&inputPath, "input", "", "JSON file containing package roots to classify")
	flag.StringVar(&outputPath, "output", "", "JSON file to write classified licenses")
	flag.Parse()
	if inputPath == "" || outputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --input and --output are required")
		os.Exit(2)
	}

	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		fail(err)
	}
	var input struct {
		Packages []resolver.PackageInput `json:"packages"`
	}
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		fail(fmt.Errorf("parse input: %w", err))
	}

	classifier, err := resolver.NewClassifier(resolver.Options{})
	if err != nil {
		fail(err)
	}
	output, err := classifier.Resolve(input.Packages)
	if err != nil {
		fail(err)
	}

	outputBytes, err := json.MarshalIndent(output, "", "  ")
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
