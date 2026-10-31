package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// writePackage writes a package_metadata file, optionally with a license
// attribute, and returns the node for it.
func writePackage(t *testing.T, dir, name, licenseJSON string) sbom.NodeConfig {
	t.Helper()
	attributes := "{}"
	if licenseJSON != "" {
		licensePath := filepath.Join(dir, name+".license.json")
		writeFile(t, licensePath, licenseJSON)
		attributes = fmt.Sprintf(`{"build.bazel.attribute.license": %q}`, licensePath)
	}
	metadataPath := filepath.Join(dir, name+".metadata.json")
	writeFile(t, metadataPath, fmt.Sprintf(
		`{"label": "@@//%s:package_metadata", "purl": "pkg:generic/%s@1.0.0", "attributes": %s}`,
		name, name, attributes,
	))
	return sbom.NodeConfig{Label: "//" + name, MetadataFile: metadataPath}
}

func TestGenerateDocumentLicenses(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "LICENSE")
	writeFile(t, textPath, "ACME license text")

	root := writePackage(t, dir, "root", "")
	single := writePackage(t, dir, "single", `{"expression":{"kind":{"identifier":"MIT","name":"MIT License"}},"label":"//single:license"}`)
	custom := writePackage(t, dir, "custom", fmt.Sprintf(
		`{"expression":{"kind":{"identifier":"LicenseRef-acme","name":"ACME License"}},"label":"//custom:license","text":%q}`, textPath))
	dual := writePackage(t, dir, "dual", `{"expression":{"operator":"OR","operands":[
		{"kind":{"identifier":"MIT","name":"MIT License"}},
		{"kind":{"identifier":"LicenseRef-acme","name":"ACME License"}}]},"label":"//dual:license"}`)
	exception := writePackage(t, dir, "exception", `{"expression":{"operator":"WITH","operands":[
		{"kind":{"identifier":"GPL-2.0-only","name":"GPL"}}],
		"exception":{"identifier":"AdditionRef-link","name":"Linking"}},"label":"//exception:license"}`)

	doc, err := GenerateDocument(
		sbom.GraphConfig{},
		sbom.Classifications{
			RootComponent: &root,
			Dependencies: sbom.DependencyNodes{
				Direct: []sbom.NodeConfig{single, custom, dual, exception},
			},
		},
		time.Unix(0, 0).UTC(),
	)
	require.NoError(t, err)

	declared := map[string]string{}
	for _, p := range doc.Packages {
		declared[p.PackageName] = p.PackageLicenseDeclared
	}
	assert.Equal(t, map[string]string{
		"root":      "NOASSERTION",
		"single":    "MIT",
		"custom":    "LicenseRef-acme",
		"dual":      "MIT OR LicenseRef-acme",
		"exception": "LicenseRef-GPL-2.0-only-WITH-link",
	}, declared)

	// The custom license is defined once, with its text, and the license
	// combined with the custom exception is defined as its own license.
	require.Len(t, doc.OtherLicenses, 2)
	assert.Equal(t, "LicenseRef-GPL-2.0-only-WITH-link", doc.OtherLicenses[0].LicenseIdentifier)
	assert.Equal(t, "GPL WITH Linking", doc.OtherLicenses[0].LicenseName)
	assert.Equal(t, "LicenseRef-acme", doc.OtherLicenses[1].LicenseIdentifier)
	assert.Equal(t, "ACME License", doc.OtherLicenses[1].LicenseName)
	assert.Equal(t, "ACME license text", doc.OtherLicenses[1].ExtractedText)
}

func TestGenerateDocumentCustomLicenseIdentifiers(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "LICENSE")
	writeFile(t, textPath, "restricted text")

	root := writePackage(t, dir, "root", "")
	// The name-only definition from the dual license is upgraded by the
	// single license that has a text.
	dual := writePackage(t, dir, "dual", `{"expression":{"operator":"OR","operands":[
		{"kind":{"identifier":"MIT","name":"MIT License"}},
		{"kind":{"identifier":"generic-restricted","name":"Restricted","custom":true}}]},"label":"//dual:license"}`)
	single := writePackage(t, dir, "single", fmt.Sprintf(
		`{"expression":{"kind":{"identifier":"generic-restricted","name":"Restricted","custom":true}},"label":"//single:license","text":%q}`, textPath))
	later := writePackage(t, dir, "later", `{"expression":{"kind":{"identifier":"GPL-3.0","name":"GPL"},"or_later":true},"label":"//later:license"}`)
	customLater := writePackage(t, dir, "customlater", `{"expression":{"kind":{"identifier":"x-y","name":"XY","custom":true},"or_later":true},"label":"//customlater:license"}`)

	doc, err := GenerateDocument(
		sbom.GraphConfig{},
		sbom.Classifications{
			RootComponent: &root,
			Dependencies:  sbom.DependencyNodes{Direct: []sbom.NodeConfig{dual, single, later, customLater}},
		},
		time.Unix(0, 0).UTC(),
	)
	require.NoError(t, err)

	declared := map[string]string{}
	for _, p := range doc.Packages {
		declared[p.PackageName] = p.PackageLicenseDeclared
	}
	assert.Equal(t, map[string]string{
		"root":        "NOASSERTION",
		"dual":        "MIT OR LicenseRef-generic-restricted",
		"single":      "LicenseRef-generic-restricted",
		"later":       "GPL-3.0+",
		"customlater": "LicenseRef-x-y-or-later",
	}, declared)

	texts := map[string]string{}
	for _, l := range doc.OtherLicenses {
		texts[l.LicenseIdentifier] = l.ExtractedText
	}
	assert.Equal(t, "restricted text", texts["LicenseRef-generic-restricted"])
	assert.Equal(t, "XY or later", texts["LicenseRef-x-y-or-later"])
}

func TestGenerateDocumentDistinctReferences(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "LICENSE")
	writeFile(t, textPath, "full terms")

	root := writePackage(t, dir, "root", "")
	upper := writePackage(t, dir, "upper", `{"expression":{"kind":{"identifier":"LicenseRef-Acme","name":"Upper"}},"label":"//upper:license"}`)
	lower := writePackage(t, dir, "lower", `{"expression":{"kind":{"identifier":"LicenseRef-acme","name":"Lower"}},"label":"//lower:license"}`)
	explicit := writePackage(t, dir, "explicit", `{"expression":{"kind":{"identifier":"LicenseRef-MIT-WITH-link","name":"Explicit"}},"label":"//explicit:license"}`)
	combined := writePackage(t, dir, "combined", fmt.Sprintf(`{"expression":{"operator":"WITH","operands":[
		{"kind":{"identifier":"MIT","name":"MIT License"}}],
		"exception":{"identifier":"AdditionRef-link","name":"Linking","custom":true}},"label":"//combined:license","text":%q}`, textPath))

	doc, err := GenerateDocument(
		sbom.GraphConfig{},
		sbom.Classifications{
			RootComponent: &root,
			Dependencies:  sbom.DependencyNodes{Direct: []sbom.NodeConfig{upper, lower, explicit, combined}},
		},
		time.Unix(0, 0).UTC(),
	)
	require.NoError(t, err)

	declared := map[string]string{}
	for _, p := range doc.Packages {
		declared[p.PackageName] = p.PackageLicenseDeclared
	}
	// SPDX ids are case-insensitive, so these must not share a definition.
	assert.NotEqual(t, strings.ToLower(declared["upper"]), strings.ToLower(declared["lower"]))
	assert.NotEqual(t, declared["explicit"], declared["combined"])

	names := map[string]string{}
	texts := map[string]string{}
	for _, l := range doc.OtherLicenses {
		names[l.LicenseIdentifier] = l.LicenseName
		texts[l.LicenseIdentifier] = l.ExtractedText
	}
	assert.Len(t, doc.OtherLicenses, 4)
	assert.Equal(t, "Upper", names[declared["upper"]])
	assert.Equal(t, "Lower", names[declared["lower"]])
	assert.Equal(t, "Explicit", names[declared["explicit"]])
	// A standalone license with a custom exception uses the supplied text.
	assert.Equal(t, "full terms", texts[declared["combined"]])
}

func TestReadBuildTimestamp_NoPath(t *testing.T) {
	got, err := readBuildTimestamp("")
	if err != nil {
		t.Fatalf("readBuildTimestamp(\"\") error = %v", err)
	}
	if !got.Equal(time.Unix(0, 0).UTC()) {
		t.Errorf("readBuildTimestamp(\"\") = %v, want the Unix epoch", got)
	}
}

func TestReadBuildTimestamp_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "volatile-status.txt")
	if err := os.WriteFile(path, []byte("BUILD_SCM_REVISION abc123\nBUILD_TIMESTAMP 1700000000\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := readBuildTimestamp(path)
	if err != nil {
		t.Fatalf("readBuildTimestamp() error = %v", err)
	}
	want := time.Unix(1700000000, 0).UTC()
	if !got.Equal(want) {
		t.Errorf("readBuildTimestamp() = %v, want %v", got, want)
	}
}

func TestReadBuildTimestamp_MissingKeyFallsBackToEpoch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "volatile-status.txt")
	if err := os.WriteFile(path, []byte("BUILD_SCM_REVISION abc123\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := readBuildTimestamp(path)
	if err != nil {
		t.Fatalf("readBuildTimestamp() error = %v", err)
	}
	if !got.Equal(time.Unix(0, 0).UTC()) {
		t.Errorf("readBuildTimestamp() = %v, want the Unix epoch", got)
	}
}

func writePackageMetadata(t *testing.T, dir, name, label, purl string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	content := `{"label": "` + label + `", "purl": "` + purl + `", "attributes": {}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return path
}

func TestGenerateDocument_RequiredFieldsPopulated(t *testing.T) {
	dir := t.TempDir()
	rootMeta := writePackageMetadata(t, dir, "root.json", "//app:root", "pkg:bazel/app@1.0.0")
	depMeta := writePackageMetadata(t, dir, "dep.json", "//lib:dep", "pkg:golang/github.com/example/dep@v1.2.3")

	graph := sbom.GraphConfig{
		RootTarget: "//app:root",
		Nodes: []sbom.NodeConfig{
			{Label: "//app:root", MetadataFile: rootMeta},
			{Label: "//lib:dep", MetadataFile: depMeta},
		},
		Edges: []sbom.EdgeConfig{
			{From: "//app:root", To: "//lib:dep", Type: "depends_on"},
		},
	}
	classifications := sbom.Classifications{
		RootComponent: &sbom.NodeConfig{Label: "//app:root", MetadataFile: rootMeta},
		Dependencies: sbom.DependencyNodes{
			Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
		},
	}

	created := time.Unix(1700000000, 0).UTC()
	doc, err := GenerateDocument(graph, classifications, created)
	if err != nil {
		t.Fatalf("GenerateDocument() error = %v", err)
	}

	if doc.DataLicense != "CC0-1.0" {
		t.Errorf("DataLicense = %q, want %q", doc.DataLicense, "CC0-1.0")
	}
	if doc.DocumentName == "" {
		t.Error("DocumentName is empty, want the root component's name")
	}
	if doc.DocumentNamespace == "" {
		t.Error("DocumentNamespace is empty, want a namespace derived from the root component")
	}
	if doc.CreationInfo == nil {
		t.Fatal("CreationInfo is nil")
	}
	if len(doc.CreationInfo.Creators) == 0 {
		t.Error("CreationInfo.Creators is empty, want at least one Tool creator")
	}
	wantCreated := "2023-11-14T22:13:20Z"
	if doc.CreationInfo.Created != wantCreated {
		t.Errorf("CreationInfo.Created = %q, want %q", doc.CreationInfo.Created, wantCreated)
	}
	for _, pkg := range doc.Packages {
		if pkg.PackageDownloadLocation == "" {
			t.Errorf("PackageDownloadLocation for %s is empty, want an SPDX sentinel value", pkg.PackageName)
		}
		if pkg.PackageLicenseConcluded != "NOASSERTION" {
			t.Errorf("PackageLicenseConcluded for %s = %q, want NOASSERTION", pkg.PackageName, pkg.PackageLicenseConcluded)
		}
		if pkg.PackageLicenseDeclared != "NOASSERTION" {
			t.Errorf("PackageLicenseDeclared for %s = %q, want NOASSERTION", pkg.PackageName, pkg.PackageLicenseDeclared)
		}
	}
}

func TestGenerateDocument_NoRootFallsBackToFirstPackage(t *testing.T) {
	dir := t.TempDir()
	depMeta := writePackageMetadata(t, dir, "dep.json", "//lib:dep", "pkg:golang/github.com/example/dep@v1.2.3")

	graph := sbom.GraphConfig{
		Nodes: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
	}
	classifications := sbom.Classifications{
		Dependencies: sbom.DependencyNodes{
			Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
		},
	}

	doc, err := GenerateDocument(graph, classifications, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument() error = %v", err)
	}
	if doc.DocumentName != "dep" {
		t.Errorf("DocumentName = %q, want %q", doc.DocumentName, "dep")
	}
}

func TestGenerateDocument_UsesPURLDownloadQualifier(t *testing.T) {
	dir := t.TempDir()
	depMeta := writePackageMetadata(t, dir, "dep.json", "//lib:dep", "pkg:golang/github.com/example/dep@v1.2.3?download_url=https://example.com/dep.tar.gz")

	graph := sbom.GraphConfig{
		Nodes: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
	}
	classifications := sbom.Classifications{
		Dependencies: sbom.DependencyNodes{
			Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
		},
	}

	doc, err := GenerateDocument(graph, classifications, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument() error = %v", err)
	}
	if got, want := doc.Packages[0].PackageDownloadLocation, "https://example.com/dep.tar.gz"; got != want {
		t.Errorf("PackageDownloadLocation = %q, want %q", got, want)
	}
}

func TestGenerateDocument_UsesPURLURLDownloadQualifier(t *testing.T) {
	dir := t.TempDir()
	depMeta := writePackageMetadata(t, dir, "dep.json", "//lib:dep", "pkg:golang/github.com/example/dep@v1.2.3?url_download=https://example.com/dep.tar.gz")

	graph := sbom.GraphConfig{
		Nodes: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
	}
	classifications := sbom.Classifications{
		Dependencies: sbom.DependencyNodes{
			Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
		},
	}

	doc, err := GenerateDocument(graph, classifications, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument() error = %v", err)
	}
	if got, want := doc.Packages[0].PackageDownloadLocation, "https://example.com/dep.tar.gz"; got != want {
		t.Errorf("PackageDownloadLocation = %q, want %q", got, want)
	}
}

func TestGenerateDocument_EmptyGraphUsesFallbackSubject(t *testing.T) {
	doc, err := GenerateDocument(sbom.GraphConfig{}, sbom.Classifications{}, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument() error = %v", err)
	}
	if doc.DocumentName == "" {
		t.Error("DocumentName is empty even with no packages at all")
	}
	if doc.DocumentNamespace == "" {
		t.Error("DocumentNamespace is empty even with no packages at all")
	}
}
