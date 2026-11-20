package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("dependencies = %v, want %v", got, want)
			}
		})
	}
}

func TestGenerateBOM_EncodesContainsAsCompositions(t *testing.T) {
	dir := t.TempDir()
	root := sbom.NodeConfig{Label: "//app:root", MetadataFile: writePackageMetadata(t, dir, "root.json", "//fixture:root", "pkg:generic/root@1.0.0")}
	child := sbom.NodeConfig{Label: "//app:child", MetadataFile: writePackageMetadata(t, dir, "child.json", "//fixture:child", "pkg:generic/child@1.0.0")}
	embedded := sbom.NodeConfig{Label: "//app:embedded", MetadataFile: writePackageMetadata(t, dir, "embedded.json", "//fixture:embedded", "pkg:generic/embedded@1.0.0")}
	build := sbom.NodeConfig{Label: "//app:build", MetadataFile: writePackageMetadata(t, dir, "build.json", "//fixture:build", "pkg:generic/build@1.0.0")}
	static := sbom.NodeConfig{Label: "//app:static", MetadataFile: writePackageMetadata(t, dir, "static.json", "//fixture:static", "pkg:generic/static@1.0.0")}
	graph := sbom.GraphConfig{
		RootTarget: root.Label,
		Nodes:      []sbom.NodeConfig{root, child, embedded, build, static},
		Edges: []sbom.EdgeConfig{
			{From: root.Label, To: child.Label, Relationship: "contains"},
			{From: child.Label, To: embedded.Label, Involvement: "embedded"},
			{From: child.Label, To: root.Label, Relationship: "contains"},
			{From: root.Label, To: build.Label, Involvement: "build", Relationship: "contains"},
			{From: root.Label, To: static.Label, Relationship: "static_link"},
		},
	}
	bom, err := GenerateBOM(graph, sbom.Classifications{
		RootComponent: &root,
		Dependencies: sbom.DependencyNodes{
			Direct: []sbom.NodeConfig{child, embedded, build, static},
		},
	})
	if err != nil {
		t.Fatalf("GenerateBOM() error = %v", err)
	}

	if bom.Compositions == nil {
		t.Fatal("Compositions is nil, want contains edges represented as compositions")
	}
	wantCompositions := []cdx.Composition{
		{
			BOMRef:     "pkg:generic/child@1.0.0",
			Aggregate:  cdx.CompositionAggregateUnknown,
			Assemblies: bomRefPointer("pkg:generic/embedded@1.0.0", "pkg:generic/root@1.0.0"),
		},
		{
			BOMRef:     "pkg:generic/root@1.0.0",
			Aggregate:  cdx.CompositionAggregateUnknown,
			Assemblies: bomRefPointer("pkg:generic/child@1.0.0"),
		},
	}
	if !reflect.DeepEqual(*bom.Compositions, wantCompositions) {
		t.Fatalf("Compositions = %#v, want %#v", *bom.Compositions, wantCompositions)
	}

	if bom.Dependencies == nil {
		t.Fatal("Dependencies is nil, want non-contains edges retained")
	}
	dependencies := make(map[string][]string)
	for _, dependency := range *bom.Dependencies {
		if dependency.Dependencies == nil {
			t.Fatalf("missing dependencies for %s", dependency.Ref)
		}
		dependencies[dependency.Ref] = *dependency.Dependencies
	}
	if got, want := dependencies["pkg:generic/root@1.0.0"], []string{
		"pkg:generic/build@1.0.0",
		"pkg:generic/static@1.0.0",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("root dependencies = %v, want %v", got, want)
	}
	if got := dependencies["pkg:generic/child@1.0.0"]; len(got) != 0 {
		t.Fatalf("child dependencies = %v, want no contains edges", got)
	}

	foundStaticLinkProperty := false
	for _, component := range append(*bom.Components, *bom.Metadata.Component) {
		if component.Properties != nil {
			for _, property := range *component.Properties {
				if property.Name == "bazel:relationship:contains" {
					t.Fatalf("contains property emitted on %s", component.BOMRef)
				}
				if component.BOMRef == "pkg:generic/root@1.0.0" &&
					property.Name == "bazel:relationship:static_link" &&
					property.Value == "pkg:generic/static@1.0.0" {
					foundStaticLinkProperty = true
				}
			}
		}
	}
	if !foundStaticLinkProperty {
		t.Fatal("static_link relationship property was not retained")
	}
}

func bomRefPointer(refs ...string) *[]cdx.BOMReference {
	bomRefs := make([]cdx.BOMReference, len(refs))
	for i, ref := range refs {
		bomRefs[i] = cdx.BOMReference(ref)
	}
	return &bomRefs
}
