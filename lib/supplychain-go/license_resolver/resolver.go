// Package license_resolver identifies SPDX licenses in fetched source trees.
//
// The resolver deliberately operates on a directory rather than a language
// package. A Bazel integration can map Go modules, Cargo crates, Maven
// artifacts, or any other fetched source to a directory and use the same
// evidence and confidence rules.
package license_resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	classifier "github.com/google/licenseclassifier/v2"
	"github.com/google/licenseclassifier/v2/assets"
)

const (
	DefaultMinimumConfidence      = 0.8
	DefaultMinimumCoveragePercent = 80.0
)

var (
	licenseFileName = regexp.MustCompile(`(?i)^(UN)?LICEN(S|C)E|COPYING|README|NOTICE.*$`)
	apache2Notice   = regexp.MustCompile(`(?im)^licensed under the apache license, version 2\.0 \(the "license"\);$`)
)

// Options controls the evidence threshold used by Classifier.
type Options struct {
	MinimumConfidence      float64
	MinimumCoveragePercent float64
}

// Evidence records one license-bearing file and the classifier's result.
type Evidence struct {
	File            string   `json:"file"`
	LicenseIDs      []string `json:"license_ids"`
	CoveragePercent float64  `json:"coverage_percent"`
}

// Result is the conservative license result for one source directory.
type Result struct {
	LicenseExpression string     `json:"license_expression,omitempty"`
	Evidence          []Evidence `json:"evidence,omitempty"`
}

// PackageInput identifies one fetched package source tree.
type PackageInput struct {
	PURL                string `json:"purl"`
	Root                string `json:"root"`
	SourceMetadataLabel string `json:"source_metadata_label,omitempty"`
}

// Response is the stable JSON contract shared by resolver integrations.
type Response struct {
	Packages []PackageResult `json:"packages"`
}

// PackageResult is the resolver output for one package source tree.
type PackageResult struct {
	PURL                string     `json:"purl"`
	SourceMetadataLabel string     `json:"source_metadata_label,omitempty"`
	LicenseExpression   string     `json:"license_expression,omitempty"`
	Evidence            []Evidence `json:"evidence,omitempty"`
}

// Classifier identifies licenses using the upstream licenseclassifier corpus.
type Classifier struct {
	classifier             *classifier.Classifier
	minimumConfidence      float64
	minimumCoveragePercent float64
}

// NewClassifier loads the SPDX corpus and returns a reusable classifier.
func NewClassifier(options Options) (*Classifier, error) {
	if options.MinimumConfidence == 0 {
		options.MinimumConfidence = DefaultMinimumConfidence
	}
	if options.MinimumCoveragePercent == 0 {
		options.MinimumCoveragePercent = DefaultMinimumCoveragePercent
	}
	if options.MinimumConfidence < 0 || options.MinimumConfidence > 1 {
		return nil, fmt.Errorf("minimum confidence must be between 0 and 1")
	}
	if options.MinimumCoveragePercent < 0 || options.MinimumCoveragePercent > 100 {
		return nil, fmt.Errorf("minimum coverage percent must be between 0 and 100")
	}

	corpus, err := assets.DefaultClassifier()
	if err != nil {
		return nil, fmt.Errorf("load upstream license corpus: %w", err)
	}
	return &Classifier{
		classifier:             corpus,
		minimumConfidence:      options.MinimumConfidence,
		minimumCoveragePercent: options.MinimumCoveragePercent,
	}, nil
}

// ClassifyDirectory finds candidate license files in root, classifies them, and
// combines all complete findings conservatively. Multiple detected licenses
// are joined with AND so a policy cannot silently discard a
// license found in a second notice or license file.
func (c *Classifier) ClassifyDirectory(root string) (Result, error) {
	candidates, err := findCandidates(root)
	if err != nil {
		return Result{}, err
	}
	candidates = primaryLicenseCandidates(candidates)

	evidence := make([]Evidence, 0, len(candidates))
	for _, candidate := range candidates {
		content, err := os.ReadFile(candidate)
		if err != nil {
			return Result{}, err
		}
		matches := c.classifier.Match(content)
		ids := make(map[string]struct{})
		for _, match := range matches.Matches {
			if match.MatchType == "License" && match.Confidence >= c.minimumConfidence && match.Name != "" {
				ids[match.Name] = struct{}{}
			}
		}

		// Some modules publish the standard Apache notice rather than the full
		// license text. It is an unambiguous declaration even without a corpus
		// match, so retain it as evidence.
		standardApacheNotice := len(ids) == 0 && isApache2Notice(candidate, content)
		if standardApacheNotice {
			ids["Apache-2.0"] = struct{}{}
		}
		if len(ids) == 0 {
			continue
		}

		licenseIDs := make([]string, 0, len(ids))
		for id := range ids {
			licenseIDs = append(licenseIDs, id)
		}
		sort.Strings(licenseIDs)
		coverage := licenseCoveragePercent(matches, c.minimumConfidence)
		if standardApacheNotice {
			coverage = 100
		}
		evidence = append(evidence, Evidence{
			File:            filepath.Base(candidate),
			LicenseIDs:      licenseIDs,
			CoveragePercent: coverage,
		})
	}

	sort.Slice(evidence, func(i, j int) bool { return evidence[i].File < evidence[j].File })
	return Result{
		LicenseExpression: conservativeExpression(evidence, c.minimumCoveragePercent),
		Evidence:          evidence,
	}, nil
}

// Resolve classifies each package source tree with one loaded corpus.
func (c *Classifier) Resolve(packages []PackageInput) (Response, error) {
	response := Response{Packages: make([]PackageResult, 0, len(packages))}
	for _, pkg := range packages {
		result, err := c.ClassifyDirectory(pkg.Root)
		if err != nil {
			return Response{}, fmt.Errorf("classify %s: %w", pkg.PURL, err)
		}
		response.Packages = append(response.Packages, PackageResult{
			PURL:                pkg.PURL,
			SourceMetadataLabel: pkg.SourceMetadataLabel,
			LicenseExpression:   result.LicenseExpression,
			Evidence:            result.Evidence,
		})
	}
	return response, nil
}

func findCandidates(root string) ([]string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("license resolver root %q is not a directory", root)
	}

	var found []string
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && licenseFileName.MatchString(entry.Name()) {
			found = append(found, filepath.Join(root, entry.Name()))
		}
	}
	return found, nil
}

func primaryLicenseCandidates(candidates []string) []string {
	primary := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		name := strings.ToUpper(filepath.Base(candidate))
		if strings.HasPrefix(name, "LICENSE") || strings.HasPrefix(name, "NOTICE") || strings.HasPrefix(name, "COPYING") {
			primary = append(primary, candidate)
		}
	}
	if len(primary) > 0 {
		return primary
	}
	return candidates
}

func isApache2Notice(candidate string, content []byte) bool {
	name := strings.ToUpper(filepath.Base(candidate))
	return (strings.HasPrefix(name, "LICENSE") || strings.HasPrefix(name, "NOTICE")) && apache2Notice.Match(content)
}

func licenseCoveragePercent(result classifier.Results, minimumConfidence float64) float64 {
	if result.TotalInputLines == 0 {
		return 0
	}
	firstLicenseLine := result.TotalInputLines + 1
	for _, match := range result.Matches {
		if match.MatchType == "License" && match.Confidence >= minimumConfidence && match.StartLine < firstLicenseLine {
			firstLicenseLine = match.StartLine
		}
	}
	if firstLicenseLine > result.TotalInputLines {
		return 0
	}
	coveredLines := make([]bool, result.TotalInputLines)
	for _, match := range result.Matches {
		if match.Confidence < minimumConfidence || (match.MatchType != "License" && match.MatchType != "Copyright") {
			continue
		}
		start := max(match.StartLine, 1)
		end := min(match.EndLine, result.TotalInputLines)
		for line := start; line <= end; line++ {
			coveredLines[line-1] = true
		}
	}
	covered := 0
	for _, lineCovered := range coveredLines[firstLicenseLine-1:] {
		if lineCovered {
			covered++
		}
	}
	return float64(covered) * 100 / float64(result.TotalInputLines-firstLicenseLine+1)
}

func conservativeExpression(evidence []Evidence, minimumCoveragePercent float64) string {
	ids := make(map[string]struct{})
	for _, item := range evidence {
		if item.CoveragePercent < minimumCoveragePercent {
			return ""
		}
		for _, id := range item.LicenseIDs {
			ids[id] = struct{}{}
		}
	}
	if len(ids) == 0 {
		return ""
	}
	terms := make([]string, 0, len(ids))
	for id := range ids {
		terms = append(terms, id)
	}
	sort.Strings(terms)
	return strings.Join(terms, " AND ")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
