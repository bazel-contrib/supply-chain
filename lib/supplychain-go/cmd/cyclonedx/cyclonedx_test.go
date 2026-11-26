package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	supplychain "github.com/bazel-contrib/supply-chain/lib/supplychain-go"
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

func TestGenerateBOM_RetainsPackageEdges(t *testing.T) {
	for _, wrappers := range []bool{false, true} {
		t.Run(fmt.Sprintf("wrappers=%t", wrappers), func(t *testing.T) {
			dir := t.TempDir()
			root := sbom.NodeConfig{Label: "//app:root", MetadataFile: writePackageMetadata(t, dir, "root.json", "//fixture:root", "pkg:golang/example.com/app@v1")}
			direct := sbom.NodeConfig{Label: "//module_a:library", MetadataFile: writePackageMetadata(t, dir, "direct.json", "//fixture:direct", "pkg:golang/example.com/module_a@v1")}
			dep := sbom.NodeConfig{Label: "//module_b:other", MetadataFile: writePackageMetadata(t, dir, "dep.json", "//fixture:dep", "pkg:golang/example.com/module_b@v1")}
			alias := sbom.NodeConfig{Label: "//module_b:used", MetadataFile: dep.MetadataFile}
			graph := sbom.GraphConfig{
				RootTarget: root.Label,
				Nodes:      []sbom.NodeConfig{root, direct, dep, alias},
				Edges: []sbom.EdgeConfig{
					{From: root.Label, To: direct.Label, Type: "depends_on"},
					{From: direct.Label, To: alias.Label, Type: "depends_on"},
					{From: alias.Label, To: dep.Label, Type: "depends_on"},
					{From: direct.Label, To: alias.Label, Type: "depends_on"},
				},
			}
			if wrappers {
				graph.Nodes = append(graph.Nodes, sbom.NodeConfig{Label: "//app:wrapper"}, sbom.NodeConfig{Label: "//module_a:wrapper"})
				graph.Edges[0].To = "//app:wrapper"
				graph.Edges[1].To = "//module_a:wrapper"
				graph.Edges[3].To = "//module_a:wrapper"
				graph.Edges = append(graph.Edges,
					sbom.EdgeConfig{From: "//app:wrapper", To: direct.Label, Type: "depends_on"},
					sbom.EdgeConfig{From: "//module_a:wrapper", To: alias.Label, Type: "depends_on"},
				)
			}
			classifications, err := sbom.ComputeClassifications(graph, false)
			if err != nil {
				t.Fatal(err)
			}
			bom, err := GenerateBOM(graph, classifications)
			if err != nil {
				t.Fatal(err)
			}
			if bom.Components == nil || len(*bom.Components) != 2 {
				t.Fatal("expected two dependency components")
			}
			if bom.Metadata == nil || bom.Metadata.Component == nil || bom.Metadata.Component.BOMRef != "pkg:golang/example.com/app@v1" {
				t.Fatal("expected app as root component")
			}
			if bom.Dependencies == nil {
				t.Fatal("dependencies is nil")
			}
			got := make(map[string][]string)
			for _, dependency := range *bom.Dependencies {
				if _, exists := got[dependency.Ref]; exists {
					t.Fatalf("duplicate dependency ref %s", dependency.Ref)
				}
				if dependency.Dependencies == nil {
					t.Fatalf("missing children for %s", dependency.Ref)
				}
				got[dependency.Ref] = *dependency.Dependencies
			}
			want := map[string][]string{
				"pkg:golang/example.com/app@v1":      {"pkg:golang/example.com/module_a@v1"},
				"pkg:golang/example.com/module_a@v1": {"pkg:golang/example.com/module_b@v1"},
				"pkg:golang/example.com/module_b@v1": {},
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("dependencies = %v, want %v", got, want)
			}
		})
	}
}

func TestHashesOf(t *testing.T) {
	got, err := hashesOf(supplychain.IntegrityAttribute{"sha256-" + strings.Repeat("A", 43) + "=", "md5-AAAA"})
	if err != nil {
		t.Fatal(err)
	}
	want := []cdx.Hash{{Algorithm: cdx.HashAlgoSHA256, Value: strings.Repeat("00", 32)}}
	if got == nil || !reflect.DeepEqual(*got, want) {
		t.Errorf("hashesOf = %v, want %v", got, want)
	}

	if got, err := hashesOf(supplychain.IntegrityAttribute{"md5-AAAA"}); err != nil || got != nil {
		t.Errorf("hashesOf(unsupported only) = %v, %v, want nil, nil", got, err)
	}
	if _, err := hashesOf(supplychain.IntegrityAttribute{"sha256-AAAA"}); err == nil {
		t.Error("hashesOf(malformed) succeeded, want an error")
	}
}

func TestMakeAcyclic_KeepsRootPathToBuildToolInCycle(t *testing.T) {
	// root -build-> tool, tool -runtime-> root.
	edges := []componentDependencyEdge{
		{from: "root", to: "tool", involvement: "build"},
		{from: "tool", to: "root", involvement: "runtime"},
	}
	got := makeAcyclicComponentDependencies(edges, []string{"root"})
	for _, e := range got {
		if e.from == "root" && e.to == "tool" {
			return
		}
	}
	t.Errorf("no path from root to tool after breaking the cycle: %v", got)
}

func TestGenerateBOM_KeepsBuildToolPathWithoutRootMetadata(t *testing.T) {
	dir := t.TempDir()
	// R has no metadata. A1 and A2 share a PURL and collapse into component A.
	a1 := writePackageMetadata(t, dir, "a1.json", "//fixture:a1", "pkg:golang/example.com/a@v1")
	a2 := writePackageMetadata(t, dir, "a2.json", "//fixture:a2", "pkg:golang/example.com/a@v1")
	b := writePackageMetadata(t, dir, "b.json", "//fixture:b", "pkg:golang/example.com/b@v1")
	graph := sbom.GraphConfig{
		RootTarget: "//app:r",
		Nodes: []sbom.NodeConfig{
			{Label: "//app:r"},
			{Label: "//a:one", MetadataFile: a1},
			{Label: "//a:two", MetadataFile: a2},
			{Label: "//b:tool", MetadataFile: b},
		},
		Edges: []sbom.EdgeConfig{
			{From: "//app:r", To: "//a:one", Type: "depends_on", Involvement: "runtime"},
			{From: "//a:one", To: "//b:tool", Type: "depends_on", Involvement: "build"},
			{From: "//b:tool", To: "//a:two", Type: "depends_on", Involvement: "runtime"},
		},
	}
	classifications, err := sbom.ComputeClassifications(graph, true)
	if err != nil {
		t.Fatal(err)
	}
	bom, err := GenerateBOM(graph, classifications)
	if err != nil {
		t.Fatal(err)
	}
	if bom.Dependencies == nil {
		t.Fatal("dependencies is nil")
	}
	for _, dependency := range *bom.Dependencies {
		if dependency.Ref == "pkg:golang/example.com/a@v1" && dependency.Dependencies != nil {
			for _, child := range *dependency.Dependencies {
				if child == "pkg:golang/example.com/b@v1" {
					return
				}
			}
		}
	}
	t.Errorf("component a has no dependency on build tool b: %v", *bom.Dependencies)
}
