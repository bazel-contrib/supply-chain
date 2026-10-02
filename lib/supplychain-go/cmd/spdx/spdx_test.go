package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

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
