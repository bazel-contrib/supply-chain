package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

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

func TestGenerateDocument23_RequiredFieldsPopulated(t *testing.T) {
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
	doc, err := GenerateDocument23(graph, classifications, created)
	if err != nil {
		t.Fatalf("GenerateDocument23() error = %v", err)
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
	wantVersions := map[string]string{"app": "1.0.0", "dep": "v1.2.3"}
	for _, pkg := range doc.Packages {
		if pkg.PackageVersion != wantVersions[pkg.PackageName] {
			t.Errorf("PackageVersion for %s = %q, want %q", pkg.PackageName, pkg.PackageVersion, wantVersions[pkg.PackageName])
		}
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

func TestGenerateDocument23_NoRootFallsBackToFirstPackage(t *testing.T) {
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

	doc, err := GenerateDocument23(graph, classifications, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument23() error = %v", err)
	}
	if doc.DocumentName != "dep" {
		t.Errorf("DocumentName = %q, want %q", doc.DocumentName, "dep")
	}
}

func TestGenerateDocument23_PURLWithoutVersionLeavesPackageVersionEmpty(t *testing.T) {
	dir := t.TempDir()
	depMeta := writePackageMetadata(t, dir, "dep.json", "//lib:dep", "pkg:golang/github.com/example/dep")

	graph := sbom.GraphConfig{
		Nodes: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
	}
	classifications := sbom.Classifications{
		Dependencies: sbom.DependencyNodes{
			Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: depMeta}},
		},
	}

	doc, err := GenerateDocument23(graph, classifications, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument23() error = %v", err)
	}
	if got := doc.Packages[0].PackageVersion; got != "" {
		t.Errorf("PackageVersion = %q, want empty", got)
	}
}

func TestGenerateDocument23_UsesPURLDownloadQualifier(t *testing.T) {
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

	doc, err := GenerateDocument23(graph, classifications, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument23() error = %v", err)
	}
	if got, want := doc.Packages[0].PackageDownloadLocation, "https://example.com/dep.tar.gz"; got != want {
		t.Errorf("PackageDownloadLocation = %q, want %q", got, want)
	}
}

func TestGenerateDocument23_UsesPURLURLDownloadQualifier(t *testing.T) {
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

	doc, err := GenerateDocument23(graph, classifications, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument23() error = %v", err)
	}
	if got, want := doc.Packages[0].PackageDownloadLocation, "https://example.com/dep.tar.gz"; got != want {
		t.Errorf("PackageDownloadLocation = %q, want %q", got, want)
	}
}

func TestGenerateDocument23_EmptyGraphUsesFallbackSubject(t *testing.T) {
	doc, err := GenerateDocument23(sbom.GraphConfig{}, sbom.Classifications{}, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument23() error = %v", err)
	}
	if doc.DocumentName == "" {
		t.Error("DocumentName is empty even with no packages at all")
	}
	if doc.DocumentNamespace == "" {
		t.Error("DocumentNamespace is empty even with no packages at all")
	}
}

func TestGenerateDocument23_RetainsPackageEdges(t *testing.T) {
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
			doc, err := GenerateDocument23(graph, classifications, time.Unix(0, 0).UTC())
			if err != nil {
				t.Fatal(err)
			}
			if len(doc.Packages) != 3 {
				t.Fatalf("got %d packages, want 3", len(doc.Packages))
			}
			got := make(map[string]int)
			for _, relationship := range doc.Relationships {
				key := string(relationship.RefA.ElementRefID) + " " + relationship.Relationship + " " + string(relationship.RefB.ElementRefID)
				got[key]++
			}
			want := map[string]int{
				"dep-0 DEPENDS_ON dep-1":   1,
				"dep-1 DEPENDS_ON dep-2":   1,
				"DOCUMENT DESCRIBES dep-0": 1,
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("relationships = %v, want %v", got, want)
			}
		})
	}
}
