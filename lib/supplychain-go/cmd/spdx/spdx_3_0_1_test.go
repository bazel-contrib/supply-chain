package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	spdx "github.com/spdx/tools-golang/spdx/v3/v3_0"
)

func TestGenerateDocument301_MapsSupplierAndOriginator(t *testing.T) {
	dir := t.TempDir()
	metadata := writeAgentMetadata301(t, dir, map[string]string{
		"build.bazel.attribute.supplier":   `{"name":"ACME Corp","type":"organization","url":"https://acme.example","email":"oss@acme.example"}`,
		"build.bazel.attribute.originator": `{"name":"A. Creator","type":"person","email":"creator@example"}`,
	})
	doc := generateAgentDocument301(t, metadata)
	pkg := agentPackage301(t, doc)

	supplier, ok := pkg.SuppliedBy.(*spdx.Organization)
	if !ok || supplier.Name != "ACME Corp" {
		t.Fatalf("SuppliedBy = %#v, want ACME Corp organization", pkg.SuppliedBy)
	}
	if len(supplier.ExternalIdentifiers) != 1 || supplier.ExternalIdentifiers[0].(*spdx.ExternalIdentifier).Identifier != "oss@acme.example" {
		t.Fatalf("supplier email = %#v", supplier.ExternalIdentifiers)
	}
	if len(supplier.ExternalRefs) != 1 || supplier.ExternalRefs[0].(*spdx.ExternalRef).Locators[0] != "https://acme.example" {
		t.Fatalf("supplier URL = %#v", supplier.ExternalRefs)
	}

	if len(pkg.OriginatedBy) != 1 {
		t.Fatalf("OriginatedBy = %#v, want one agent", pkg.OriginatedBy)
	}
	originator, ok := pkg.OriginatedBy[0].(*spdx.Person)
	if !ok || originator.Name != "A. Creator" {
		t.Fatalf("OriginatedBy = %#v, want A. Creator person", pkg.OriginatedBy[0])
	}
	if len(originator.ExternalIdentifiers) != 1 || originator.ExternalIdentifiers[0].(*spdx.ExternalIdentifier).Identifier != "creator@example" {
		t.Fatalf("originator email = %#v", originator.ExternalIdentifiers)
	}

	element := doc.RootElements[0].(*spdx.SBOM)
	if len(element.Elements) != 3 {
		t.Fatalf("SBOM Elements = %d, want package plus two agents", len(element.Elements))
	}
}

func TestGenerateDocument301_AbsentAgentsAreOmitted(t *testing.T) {
	dir := t.TempDir()
	metadata := writeAgentMetadata301(t, dir, nil)
	doc := generateAgentDocument301(t, metadata)
	pkg := agentPackage301(t, doc)
	if pkg.SuppliedBy != nil || len(pkg.OriginatedBy) != 0 {
		t.Fatalf("package agent fields are populated: supplied=%#v originated=%#v", pkg.SuppliedBy, pkg.OriginatedBy)
	}
	if got := len(doc.RootElements[0].(*spdx.SBOM).Elements); got != 1 {
		t.Fatalf("SBOM Elements = %d, want only the package", got)
	}
}

func TestGenerateDocument301_MapsPersonSupplier(t *testing.T) {
	dir := t.TempDir()
	metadata := writeAgentMetadata301(t, dir, map[string]string{
		"build.bazel.attribute.supplier": `{"name":"A. Supplier","type":"person","email":"supplier@example"}`,
	})
	doc := generateAgentDocument301(t, metadata)
	pkg := agentPackage301(t, doc)
	supplier, ok := pkg.SuppliedBy.(*spdx.Person)
	if !ok || supplier.Name != "A. Supplier" {
		t.Fatalf("SuppliedBy = %#v, want A. Supplier person", pkg.SuppliedBy)
	}
	if len(supplier.ExternalIdentifiers) != 1 || supplier.ExternalIdentifiers[0].(*spdx.ExternalIdentifier).Identifier != "supplier@example" {
		t.Fatalf("supplier email = %#v", supplier.ExternalIdentifiers)
	}
	if len(pkg.OriginatedBy) != 0 {
		t.Fatalf("OriginatedBy = %#v, want empty", pkg.OriginatedBy)
	}
}

func TestGenerateDocument301_DeduplicatesIdenticalAgents(t *testing.T) {
	dir := t.TempDir()
	metadata := writeAgentMetadata301(t, dir, map[string]string{
		"build.bazel.attribute.supplier":   `{"name":"Same Entity","type":"organization"}`,
		"build.bazel.attribute.originator": `{"name":"Same Entity","type":"organization"}`,
	})
	doc := generateAgentDocument301(t, metadata)
	pkg := agentPackage301(t, doc)
	if len(pkg.OriginatedBy) != 1 || pkg.SuppliedBy != pkg.OriginatedBy[0] {
		t.Fatalf("agent references differ: supplier=%#v originator=%#v", pkg.SuppliedBy, pkg.OriginatedBy)
	}
	if got := len(doc.RootElements[0].(*spdx.SBOM).Elements); got != 2 {
		t.Fatalf("SBOM Elements = %d, want package plus one deduplicated agent", got)
	}
}

func generateAgentDocument301(t *testing.T, metadata string) *spdx.Document {
	t.Helper()
	doc, err := GenerateDocument301(
		sbom.GraphConfig{},
		sbom.Classifications{Dependencies: sbom.DependencyNodes{
			Direct: []sbom.NodeConfig{{Label: "//lib:dep", MetadataFile: metadata}},
		}},
		time.Unix(0, 0).UTC(),
	)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func agentPackage301(t *testing.T, doc *spdx.Document) *spdx.Package {
	t.Helper()
	element, ok := doc.RootElements[0].(*spdx.SBOM)
	if !ok || len(element.RootElements) != 0 || len(element.Elements) == 0 {
		t.Fatalf("unexpected SBOM root: %#v", doc.RootElements)
	}
	for _, entry := range element.Elements {
		if pkg, ok := entry.(*spdx.Package); ok {
			return pkg
		}
	}
	t.Fatal("package element not found")
	return nil
}

func writeAgentMetadata301(t *testing.T, dir string, agents map[string]string) string {
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
