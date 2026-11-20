package main

import (
	"encoding/json"
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
	if supplier := (*bom.Components)[0].Supplier; supplier != nil {
		t.Fatalf("Supplier = %#v, want nil when no supplier is declared", supplier)
	}
	if publisher := (*bom.Components)[0].Publisher; publisher != "" {
		t.Fatalf("Publisher = %q, want empty when no originator is declared", publisher)
	}
}

func TestGenerateBOM_MapsSupplierAndOriginator(t *testing.T) {
	dir := t.TempDir()
	metadata := writeAgentMetadata(t, dir, map[string]string{
		"build.bazel.attribute.supplier":   `{"name":"ACME Corp","type":"organization","url":"https://acme.example","email":"oss@acme.example"}`,
		"build.bazel.attribute.originator": `{"name":"A. Creator","type":"person","email":"creator@example"}`,
	})
	bom, err := GenerateBOM(
		sbom.GraphConfig{},
		sbom.Classifications{Dependencies: sbom.DependencyNodes{
			Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: metadata}},
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	component := (*bom.Components)[0]
	if component.Supplier == nil || component.Supplier.Name != "ACME Corp" {
		t.Fatalf("Supplier = %#v, want ACME Corp", component.Supplier)
	}
	if component.Supplier.URL == nil || !reflect.DeepEqual(*component.Supplier.URL, []string{"https://acme.example"}) {
		t.Fatalf("Supplier.URL = %#v", component.Supplier.URL)
	}
	if component.Supplier.Contact == nil || len(*component.Supplier.Contact) != 1 || (*component.Supplier.Contact)[0].Email != "oss@acme.example" {
		t.Fatalf("Supplier.Contact = %#v", component.Supplier.Contact)
	}
	if component.Publisher != "A. Creator" {
		t.Fatalf("Publisher = %q, want originator name", component.Publisher)
	}
}

func TestGenerateBOM_MapsPersonSupplier(t *testing.T) {
	dir := t.TempDir()
	metadata := writeAgentMetadata(t, dir, map[string]string{
		"build.bazel.attribute.supplier": `{"name":"A. Supplier","type":"person","email":"supplier@example"}`,
	})
	bom, err := GenerateBOM(
		sbom.GraphConfig{},
		sbom.Classifications{Dependencies: sbom.DependencyNodes{
			Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: metadata}},
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	supplier := (*bom.Components)[0].Supplier
	if supplier == nil || supplier.Name != "A. Supplier" || supplier.Contact == nil || (*supplier.Contact)[0].Email != "supplier@example" {
		t.Fatalf("Supplier = %#v", supplier)
	}
}

func writeAgentMetadata(t *testing.T, dir string, agents map[string]string) string {
	t.Helper()
	attributes := make(map[string]string, len(agents))
	for kind, content := range agents {
		path := filepath.Join(dir, strings.TrimPrefix(kind, "build.bazel.attribute.")+".json")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		attributes[kind] = path
	}
	data, err := json.Marshal(map[string]any{
		"label":      "//lib:dep",
		"purl":       "pkg:generic/dep@1.0",
		"attributes": attributes,
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "metadata.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
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
