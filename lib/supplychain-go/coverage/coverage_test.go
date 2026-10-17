package coverage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildReportCombinesMetadataAndResolverResults(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	licenseAttribute := filepath.Join(root, "license.attribute.json")
	metadata := filepath.Join(root, "metadata.json")
	if err := os.WriteFile(licenseAttribute, []byte(`{"kind":{"identifier":"MIT"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	metadataBytes, err := json.Marshal(map[string]any{
		"label":      "//app:root",
		"purl":       "pkg:generic/app@1",
		"attributes": map[string]string{licenseAttributeName(): licenseAttribute},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metadata, metadataBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := BuildReport(GraphConfig{
		SchemaVersion: "1.0",
		RootTarget:    "//app:root",
		Nodes: []NodeConfig{
			{Label: "//app:root", MetadataFile: metadata},
			{Label: "@@repo//dep:package", MetadataFile: metadata},
		},
	}, map[string][]string{
		"pkg:generic/app@1": {"Apache-2.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := report.UniquePackages, 1; got != want {
		t.Fatalf("got %d unique packages, want %d", got, want)
	}
	if got := report.MissingLicensePURLs; len(got) != 0 {
		t.Fatalf("got missing packages %v", got)
	}
	if got, want := report.LicensedPackages["pkg:generic/app@1"], []string{"Apache-2.0", "MIT"}; !equalStrings(got, want) {
		t.Fatalf("got identifiers %v, want %v", got, want)
	}
}

func licenseAttributeName() string {
	return "build.bazel.attribute.license"
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
