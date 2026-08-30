// Package coverage builds a language-neutral license metadata report from the
// graph emitted by tools/gather_metadata. Language-specific resolvers can add
// identifiers for packages whose Bazel metadata does not include a license.
package coverage

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
)

const licenseAttribute = "build.bazel.attribute.license"

type licenseAttributeValue struct {
	Kind struct {
		Identifier string `json:"identifier"`
	} `json:"kind"`
}

var licenseAttributeDescriptor = supplychain.PackageAttributeDescriptor[licenseAttributeValue]{
	Kind: licenseAttribute,
	Parser: func(reader io.Reader) (*licenseAttributeValue, error) {
		var attribute licenseAttributeValue
		if err := json.NewDecoder(reader).Decode(&attribute); err != nil {
			return nil, err
		}
		return &attribute, nil
	},
}

// GraphConfig is the graph-only JSON emitted by tools/gather_metadata.
type GraphConfig struct {
	SchemaVersion string       `json:"schema_version"`
	RootTarget    string       `json:"root_target"`
	Nodes         []NodeConfig `json:"nodes"`
	Edges         []EdgeConfig `json:"edges"`
}

type NodeConfig struct {
	Label        string `json:"label"`
	MetadataFile string `json:"metadata_file"`
}

type EdgeConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// Report is the stable input contract for policy evaluators.
type Report struct {
	SchemaVersion                  int                 `json:"schema_version"`
	RootTarget                     string              `json:"root_target,omitempty"`
	MetadataRecords                int                 `json:"metadata_records"`
	UniqueMetadataRecords          int                 `json:"unique_metadata_records"`
	DuplicateMetadataRecords       int                 `json:"duplicate_metadata_records"`
	RecordsWithoutPURL             int                 `json:"records_without_purl"`
	UniquePackages                 int                 `json:"unique_packages"`
	PackagesWithLicenseMetadata    int                 `json:"packages_with_license_metadata"`
	LicenseMetadataCoveragePercent float64             `json:"license_metadata_coverage_percent"`
	LicensedPackages               map[string][]string `json:"licensed_packages"`
	MissingLicensePURLs            []string            `json:"missing_license_purls"`
}

// LicenseExpression returns the SPDX expression recorded for purl. Multiple
// independent findings are combined conservatively with AND.
func (r Report) LicenseExpression(purl string) string {
	expressions := r.LicensedPackages[purl]
	if len(expressions) == 0 {
		return ""
	}
	return strings.Join(expressions, " AND ")
}

// BuildReport reads each unique metadata file exactly once. The resolver map
// is keyed by PURL and can contain results from Go, Rust, Java, or any other
// language-specific implementation.
func BuildReport(graph GraphConfig, resolved map[string][]string) (Report, error) {
	paths := make([]string, 0, len(graph.Nodes))
	seenPaths := make(map[string]bool)
	for _, node := range graph.Nodes {
		if node.MetadataFile == "" || seenPaths[node.MetadataFile] {
			continue
		}
		seenPaths[node.MetadataFile] = true
		paths = append(paths, node.MetadataFile)
	}
	sort.Strings(paths)

	type packageRecord struct {
		labels      map[string]bool
		identifiers map[string]bool
	}
	packages := make(map[string]*packageRecord)
	recordsWithoutPURL := 0
	for _, path := range paths {
		metadata, err := supplychain.ReadPackageMetadataFromFile(path)
		if err != nil {
			return Report{}, fmt.Errorf("read package metadata %q: %w", path, err)
		}
		purl := metadata.GetPURL().String()
		if purl == "" {
			recordsWithoutPURL++
			continue
		}
		record := packages[purl]
		if record == nil {
			record = &packageRecord{labels: make(map[string]bool), identifiers: make(map[string]bool)}
			packages[purl] = record
		}
		record.labels[metadata.GetLabel().String()] = true
		identifier, err := licenseIdentifier(metadata)
		if err != nil {
			return Report{}, fmt.Errorf("read license metadata for %q: %w", purl, err)
		}
		if identifier != "" {
			record.identifiers[identifier] = true
		}
	}
	for purl, identifiers := range resolved {
		if record := packages[purl]; record != nil {
			for _, identifier := range identifiers {
				if identifier != "" {
					record.identifiers[identifier] = true
				}
			}
		}
	}

	licensed := make(map[string][]string)
	missing := make([]string, 0)
	for purl, record := range packages {
		identifiers := sortedKeys(record.identifiers)
		if len(identifiers) == 0 {
			missing = append(missing, purl)
			continue
		}
		licensed[purl] = identifiers
	}
	sort.Strings(missing)
	coveragePercent := 100.0
	if len(packages) > 0 {
		coveragePercent = float64(len(licensed)) * 100 / float64(len(packages))
	}

	report := Report{
		SchemaVersion:                  1,
		MetadataRecords:                len(graph.Nodes),
		UniqueMetadataRecords:          len(paths),
		DuplicateMetadataRecords:       len(graph.Nodes) - len(paths),
		RecordsWithoutPURL:             recordsWithoutPURL,
		UniquePackages:                 len(packages),
		PackagesWithLicenseMetadata:    len(licensed),
		LicenseMetadataCoveragePercent: coveragePercent,
		LicensedPackages:               licensed,
		MissingLicensePURLs:            missing,
	}
	if graph.RootTarget != "" {
		report.RootTarget = graph.RootTarget
	}
	return report, nil
}

func licenseIdentifier(metadata supplychain.PackageMetadata) (string, error) {
	attributes := metadata.ListAttributeKinds()
	found := false
	for _, attribute := range attributes {
		if attribute == licenseAttribute {
			found = true
			break
		}
	}
	if !found {
		return "", nil
	}
	attribute, err := supplychain.GetPackageAttribute(metadata, licenseAttributeDescriptor)
	if err != nil {
		return "", err
	}
	return attribute.Kind.Identifier, nil
}

func sortedKeys(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
