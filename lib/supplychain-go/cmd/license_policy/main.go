package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/policy"
)

func main() {
	var reportPath, allowlistPath, outputPath string
	flag.StringVar(&reportPath, "report", "", "Coverage report JSON")
	flag.StringVar(&allowlistPath, "allowlist", "", "License policy JSON")
	flag.StringVar(&outputPath, "output", "", "Optional evaluation JSON output")
	flag.Parse()
	if reportPath == "" || allowlistPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --report and --allowlist are required")
		os.Exit(2)
	}

	allowlist, err := policy.LoadAllowlist(allowlistPath)
	if err != nil {
		fail(err)
	}
	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		fail(err)
	}
	var report policy.Report
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		fail(fmt.Errorf("parse report: %w", err))
	}
	evaluation, err := policy.Evaluate(report, allowlist, time.Now().UTC())
	if err != nil {
		fail(err)
	}

	outputBytes, err := json.MarshalIndent(evaluation, "", "  ")
	if err != nil {
		fail(err)
	}
	outputBytes = append(outputBytes, '\n')
	if outputPath != "" {
		if err := os.WriteFile(outputPath, outputBytes, 0o644); err != nil {
			fail(err)
		}
	} else {
		_, _ = os.Stdout.Write(outputBytes)
	}
	if !evaluation.Acceptable {
		os.Exit(1)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
