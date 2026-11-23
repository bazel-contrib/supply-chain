package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

func writeMetadataWithLicense(t *testing.T, dir, name, purl string, withLicense bool) string {
	t.Helper()
	attrs := `{}`
	if withLicense {
		lic := filepath.Join(dir, name+".license.json")
		if err := os.WriteFile(lic, []byte(`{"kind":{"identifier":"MIT","name":"MIT"}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		attrs = `{"build.bazel.attribute.license": "` + lic + `"}`
	}
	path := filepath.Join(dir, name)
	content := `{"label": "//x:y", "purl": "` + purl + `", "attributes": ` + attrs + `}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func duplicatePURLInputs(t *testing.T) (sbom.GraphConfig, sbom.Classifications) {
	dir := t.TempDir()
	root := writeMetadataWithLicense(t, dir, "root.json", "pkg:bazel/app@1", false)
	first := writeMetadataWithLicense(t, dir, "first.json", "pkg:golang/example.com/dep@v1", false)
	second := writeMetadataWithLicense(t, dir, "second.json", "pkg:golang/example.com/dep@v1", true)
	graph := sbom.GraphConfig{RootTarget: "//app:root"}
	cls := sbom.Classifications{
		RootComponent: &sbom.NodeConfig{Label: "//app:root", MetadataFile: root},
		Dependencies: sbom.DependencyNodes{Direct: []sbom.NodeConfig{
			{Label: "//a:dep", MetadataFile: first},
			{Label: "//b:dep", MetadataFile: second},
		}},
	}
	return graph, cls
}

func TestDuplicatePURL_LicenseOfLaterMetadataIsKept_23(t *testing.T) {
	graph, cls := duplicatePURLInputs(t)
	doc, err := GenerateDocument23(graph, cls, time.Unix(1700000000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range doc.Packages {
		if p.PackageName == "dep" {
			if p.PackageLicenseDeclared != "MIT" {
				t.Errorf("PackageLicenseDeclared = %q, want MIT", p.PackageLicenseDeclared)
			}
			return
		}
	}
	t.Fatal("dep package not found")
}

func TestDuplicatePURL_LicenseOfLaterMetadataIsKept_301(t *testing.T) {
	graph, cls := duplicatePURLInputs(t)
	doc, err := GenerateDocument301(graph, cls, time.Unix(1700000000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "hasDeclaredLicense") {
		t.Errorf("no hasDeclaredLicense relationship in document: %s", out)
	}
}
