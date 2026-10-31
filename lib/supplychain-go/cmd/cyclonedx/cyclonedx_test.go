package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writePackage(t *testing.T, dir, name, licenseJSON string) sbom.NodeConfig {
	t.Helper()
	attributes := "{}"
	if licenseJSON != "" {
		licensePath := filepath.Join(dir, name+".license.json")
		require.NoError(t, os.WriteFile(licensePath, []byte(licenseJSON), 0o644))
		attributes = fmt.Sprintf(`{"build.bazel.attribute.license": %q}`, licensePath)
	}
	metadataPath := filepath.Join(dir, name+".metadata.json")
	require.NoError(t, os.WriteFile(metadataPath, []byte(fmt.Sprintf(
		`{"label": "@@//%s:package_metadata", "purl": "pkg:generic/%s@1.0.0", "attributes": %s}`,
		name, name, attributes,
	)), 0o644))
	return sbom.NodeConfig{Label: "//" + name, MetadataFile: metadataPath}
}

func TestGenerateBOMLicenses(t *testing.T) {
	dir := t.TempDir()

	none := writePackage(t, dir, "none", "")
	single := writePackage(t, dir, "single", `{"expression":{"kind":{"identifier":"MIT","name":"MIT License"}},"label":"//single:license"}`)
	deprecated := writePackage(t, dir, "deprecated", `{"expression":{"kind":{"identifier":"GPL-2.0+","name":"GPL 2.0 or later"}},"label":"//deprecated:license"}`)
	custom := writePackage(t, dir, "custom", `{"expression":{"kind":{"identifier":"LicenseRef-acme","name":"ACME License"}},"label":"//custom:license"}`)
	dual := writePackage(t, dir, "dual", `{"expression":{"operator":"OR","operands":[
		{"kind":{"identifier":"Apache-2.0","name":"Apache"}},
		{"kind":{"identifier":"MIT","name":"MIT License"}}]},"label":"//dual:license"}`)
	exception := writePackage(t, dir, "exception", `{"expression":{"operator":"OR","operands":[
		{"kind":{"identifier":"LicenseRef-acme","name":"ACME License"}},
		{"operator":"WITH","operands":[{"kind":{"identifier":"GPL-2.0-only","name":"GPL"}}],
		 "exception":{"identifier":"Classpath-exception-2.0","name":"Classpath"}}]},"label":"//exception:license"}`)

	bom, err := GenerateBOM(
		sbom.GraphConfig{},
		sbom.Classifications{
			Dependencies: sbom.DependencyNodes{
				Direct: []sbom.NodeConfig{none, single, deprecated, custom, dual, exception},
			},
		},
	)
	require.NoError(t, err)

	licenses := map[string]*cdx.Licenses{}
	for _, c := range *bom.Components {
		licenses[c.Name] = c.Licenses
	}

	assert.Nil(t, licenses["none"])
	assert.Equal(t, &cdx.Licenses{{License: &cdx.License{ID: "MIT"}}}, licenses["single"])
	assert.Equal(t, &cdx.Licenses{{License: &cdx.License{ID: "GPL-2.0+"}}}, licenses["deprecated"])
	assert.Equal(t, &cdx.Licenses{{License: &cdx.License{Name: "ACME License"}}}, licenses["custom"])
	assert.Equal(t, &cdx.Licenses{{Expression: "Apache-2.0 OR MIT"}}, licenses["dual"])
	assert.Equal(t,
		&cdx.Licenses{{Expression: "LicenseRef-acme OR GPL-2.0-only WITH Classpath-exception-2.0"}},
		licenses["exception"],
	)
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

func TestGenerateBOM_UsesPURLDownloadQualifier(t *testing.T) {
	dir := t.TempDir()
	depMeta := writePackageMetadata(t, dir, "dep.json", "//lib:dep", "pkg:golang/github.com/example/dep@v1.2.3?download_url=https://example.com/dep.tar.gz")

	bom, err := GenerateBOM(
		sbom.GraphConfig{
			Nodes: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
		},
		sbom.Classifications{
			Dependencies: sbom.DependencyNodes{
				Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
			},
		},
	)
	if err != nil {
		t.Fatalf("GenerateBOM() error = %v", err)
	}

	assertDistributionReference(t, (*bom.Components)[0], "https://example.com/dep.tar.gz")
}

func TestGenerateBOM_UsesPURLURLDownloadQualifier(t *testing.T) {
	dir := t.TempDir()
	depMeta := writePackageMetadata(t, dir, "dep.json", "//lib:dep", "pkg:golang/github.com/example/dep@v1.2.3?url_download=https://example.com/dep.tar.gz")

	bom, err := GenerateBOM(
		sbom.GraphConfig{
			Nodes: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
		},
		sbom.Classifications{
			Dependencies: sbom.DependencyNodes{
				Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
			},
		},
	)
	if err != nil {
		t.Fatalf("GenerateBOM() error = %v", err)
	}

	assertDistributionReference(t, (*bom.Components)[0], "https://example.com/dep.tar.gz")
}

func TestGenerateBOM_WithoutDownloadQualifierOmitsDistributionReference(t *testing.T) {
	dir := t.TempDir()
	depMeta := writePackageMetadata(t, dir, "dep.json", "//lib:dep", "pkg:golang/github.com/example/dep@v1.2.3")

	bom, err := GenerateBOM(
		sbom.GraphConfig{
			Nodes: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
		},
		sbom.Classifications{
			Dependencies: sbom.DependencyNodes{
				Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
			},
		},
	)
	if err != nil {
		t.Fatalf("GenerateBOM() error = %v", err)
	}

	if refs := (*bom.Components)[0].ExternalReferences; refs != nil {
		t.Fatalf("ExternalReferences = %#v, want nil", *refs)
	}
	if licenses := (*bom.Components)[0].Licenses; licenses != nil {
		t.Fatalf("Licenses = %#v, want nil when no component license is known", *licenses)
	}
}

func assertDistributionReference(t *testing.T, component cdx.Component, wantURL string) {
	t.Helper()
	if component.ExternalReferences == nil {
		t.Fatal("ExternalReferences is nil, want distribution reference")
	}
	for _, ref := range *component.ExternalReferences {
		if ref.Type == cdx.ERTypeDistribution {
			if ref.URL != wantURL {
				t.Fatalf("distribution external reference URL = %q, want %q", ref.URL, wantURL)
			}
			return
		}
	}
	t.Fatalf("ExternalReferences = %#v, want distribution reference", *component.ExternalReferences)
}
