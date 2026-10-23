package main

import (
	"os"
	"path/filepath"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

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
