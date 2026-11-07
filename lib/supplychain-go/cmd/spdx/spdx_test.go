package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	spdx "github.com/spdx/tools-golang/spdx/v3/v3_0"
)

func sbomOf(doc *spdx.Document) *spdx.SBOM {
	return doc.RootElements[0].(*spdx.SBOM)
}

func packagesOf(doc *spdx.Document) []*spdx.Package {
	var packages []*spdx.Package
	for _, element := range sbomOf(doc).Elements {
		if pkg, ok := element.(*spdx.Package); ok {
			packages = append(packages, pkg)
		}
	}
	return packages
}

func relationshipsOf(doc *spdx.Document) []*spdx.Relationship {
	var relationships []*spdx.Relationship
	for _, element := range sbomOf(doc).Elements {
		if relationship, ok := element.(*spdx.Relationship); ok {
			relationships = append(relationships, relationship)
		}
	}
	return relationships
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

	if got := doc.DataLicense.(*spdx.ListedLicense).Name; got != "CC0-1.0" {
		t.Errorf("DataLicense = %q, want %q", got, "CC0-1.0")
	}
	if doc.Name != "app" {
		t.Errorf("Name = %q, want the root component's name", doc.Name)
	}
	if want := "https://spdx.org/spdxdocs/pkg:bazel%2Fapp@1.0.0"; doc.ID != want {
		t.Errorf("ID = %q, want %q", doc.ID, want)
	}
	if got := doc.CreationInfo.(*spdx.CreationInfo).Created; !got.Equal(created) {
		t.Errorf("Created = %v, want %v", got, created)
	}
	if got := doc.CreationInfo.(*spdx.CreationInfo).SpecVersion; got != "3.0.1" {
		t.Errorf("SpecVersion = %q, want 3.0.1", got)
	}
	wantVersions := map[string]string{"app": "1.0.0", "dep": "v1.2.3"}
	for _, pkg := range packagesOf(doc) {
		if pkg.Version != wantVersions[pkg.Name] {
			t.Errorf("Version for %s = %q, want %q", pkg.Name, pkg.Version, wantVersions[pkg.Name])
		}
		if pkg.PackageURL == "" {
			t.Errorf("PackageURL for %s is empty", pkg.Name)
		}
	}
	if roots := sbomOf(doc).RootElements; len(roots) != 1 || roots[0].(*spdx.Package).Name != "app" {
		t.Errorf("SBOM root elements = %v, want the root component", roots)
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
	if doc.Name != "dep" {
		t.Errorf("Name = %q, want %q", doc.Name, "dep")
	}
}

func TestGenerateDocument_PURLWithoutVersionLeavesPackageVersionEmpty(t *testing.T) {
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

	doc, err := GenerateDocument(graph, classifications, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument() error = %v", err)
	}
	if got := packagesOf(doc)[0].Version; got != "" {
		t.Errorf("PackageVersion = %q, want empty", got)
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
	if got, want := string(packagesOf(doc)[0].DownloadLocation), "https://example.com/dep.tar.gz"; got != want {
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
	if got, want := string(packagesOf(doc)[0].DownloadLocation), "https://example.com/dep.tar.gz"; got != want {
		t.Errorf("PackageDownloadLocation = %q, want %q", got, want)
	}
}

func TestGenerateDocument_EmptyGraphUsesFallbackSubject(t *testing.T) {
	doc, err := GenerateDocument(sbom.GraphConfig{}, sbom.Classifications{}, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument() error = %v", err)
	}
	if doc.Name == "" {
		t.Error("Name is empty even with no packages at all")
	}
	if doc.ID == "" {
		t.Error("ID is empty even with no packages at all")
	}
}

func TestGenerateDocument_RetainsPackageEdges(t *testing.T) {
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
			doc, err := GenerateDocument(graph, classifications, time.Unix(0, 0).UTC())
			if err != nil {
				t.Fatal(err)
			}
			if len(packagesOf(doc)) != 3 {
				t.Fatalf("got %d packages, want 3", len(packagesOf(doc)))
			}
			got := make(map[string]int)
			for _, relationship := range relationshipsOf(doc) {
				key := relationship.From.(*spdx.Package).ID + " " + path.Base(relationship.Type.GetID()) + " " + relationship.To[0].(*spdx.Package).ID
				got[key]++
			}
			want := map[string]int{
				"dep-0 dependsOn dep-1": 1,
				"dep-1 dependsOn dep-2": 1,
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("relationships = %v, want %v", got, want)
			}
		})
	}
}

func TestGenerateDocument_OutputIsDeterministicSPDX3JSONLD(t *testing.T) {
	dir := t.TempDir()
	rootMeta := writePackageMetadata(t, dir, "root.json", "//app:root", "pkg:bazel/app@1.0.0")
	var nodes []sbom.NodeConfig
	var edges []sbom.EdgeConfig
	classifications := sbom.Classifications{RootComponent: &sbom.NodeConfig{Label: "//app:root", MetadataFile: rootMeta}}
	nodes = append(nodes, *classifications.RootComponent)
	for i := 0; i < 20; i++ {
		label := fmt.Sprintf("//lib:dep%d", i)
		node := sbom.NodeConfig{Label: label, MetadataFile: writePackageMetadata(t, dir, fmt.Sprintf("dep%d.json", i), label, fmt.Sprintf("pkg:golang/github.com/example/dep%d@v1.2.3", i))}
		nodes = append(nodes, node)
		edges = append(edges, sbom.EdgeConfig{From: "//app:root", To: label, Type: "depends_on"})
		classifications.Dependencies.Direct = append(classifications.Dependencies.Direct, node)
	}
	graph := sbom.GraphConfig{Nodes: nodes, Edges: edges}

	render := func() string {
		doc, err := GenerateDocument(graph, classifications, time.Unix(1700000000, 0).UTC())
		if err != nil {
			t.Fatal(err)
		}
		var buf strings.Builder
		if err := writeDocument(doc, &buf); err != nil {
			t.Fatal(err)
		}
		return buf.String()
	}
	first := render()
	for i := 0; i < 5; i++ {
		if render() != first {
			t.Fatal("output differs between two runs")
		}
	}
	for _, want := range []string{`"specVersion": "3.0.1"`, `"software_Package"`, `"dependsOn"`, `"rootElement"`, `"software_packageUrl": "pkg:golang/github.com/example/dep3@v1.2.3"`} {
		if !strings.Contains(first, want) {
			t.Errorf("output does not contain %s:\n%s", want, first)
		}
	}
}

func writeLicensedPackage(t *testing.T, dir, name, licenseJSON string) sbom.NodeConfig {
	t.Helper()
	licensePath := filepath.Join(dir, name+".license.json")
	if err := os.WriteFile(licensePath, []byte(licenseJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name+".metadata.json")
	content := fmt.Sprintf(`{"label": "//fixture:%s", "purl": "pkg:generic/%s@1.0.0", "attributes": {"build.bazel.attribute.license": %q}}`, name, name, licensePath)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return sbom.NodeConfig{Label: "//" + name, MetadataFile: path}
}

func TestGenerateDocument_DeclaredLicenses(t *testing.T) {
	dir := t.TempDir()
	mit := func(name string) sbom.NodeConfig {
		return writeLicensedPackage(t, dir, name, `{"expression":{"kind":{"identifier":"MIT","name":"MIT License"}},"label":"//x:license"}`)
	}
	custom := func(name, text string) sbom.NodeConfig {
		textPath := filepath.Join(dir, name+".LICENSE")
		if err := os.WriteFile(textPath, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		return writeLicensedPackage(t, dir, name, `{"expression":{"kind":{"identifier":"LicenseRef-acme","name":"ACME License"}},"label":"//x:license","text":"`+textPath+`"}`)
	}
	legacy := writeLicensedPackage(t, dir, "f", `{"expression":{"kind":{"identifier":"internal","name":"Internal","custom":true}},"label":"//x:license"}`)
	nodes := []sbom.NodeConfig{
		mit("a"), mit("b"), legacy,
		custom("c", "Copyright 2024 ACME"), custom("d", "Copyright 2024 ACME"), custom("e", "Copyright 2025 ACME"),
	}

	doc, err := GenerateDocument(sbom.GraphConfig{}, sbom.Classifications{
		Dependencies: sbom.DependencyNodes{Direct: nodes},
	}, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument() error = %v", err)
	}

	var listed, custom_ []string
	for _, element := range sbomOf(doc).Elements {
		switch license := element.(type) {
		case *spdx.ListedLicense:
			listed = append(listed, license.ID)
		case *spdx.CustomLicense:
			custom_ = append(custom_, license.ID+"="+license.Text)
		}
	}
	if want := []string{"https://spdx.org/licenses/MIT"}; !reflect.DeepEqual(listed, want) {
		t.Errorf("listed licenses = %v, want %v", listed, want)
	}
	wantCustom := []string{"LicenseRef-internal=", "LicenseRef-acme=Copyright 2024 ACME", "LicenseRef-acme-1=Copyright 2025 ACME"}
	if !reflect.DeepEqual(custom_, wantCustom) {
		t.Errorf("custom licenses = %v, want %v", custom_, wantCustom)
	}

	declared := map[string]string{}
	for _, relationship := range relationshipsOf(doc) {
		if relationship.Type != spdx.RelationshipType_HasDeclaredLicense {
			continue
		}
		license := relationship.To[0]
		switch l := license.(type) {
		case *spdx.ListedLicense:
			declared[relationship.From.(*spdx.Package).Name] = l.Name
		case *spdx.CustomLicense:
			declared[relationship.From.(*spdx.Package).Name] = l.ID
		}
	}
	wantDeclared := map[string]string{"a": "MIT", "b": "MIT", "c": "LicenseRef-acme", "d": "LicenseRef-acme", "e": "LicenseRef-acme-1", "f": "LicenseRef-internal"}
	if !reflect.DeepEqual(declared, wantDeclared) {
		t.Errorf("declared licenses = %v, want %v", declared, wantDeclared)
	}
}

func TestGenerateDocument_LicenseExpressions(t *testing.T) {
	dir := t.TempDir()
	text := filepath.Join(dir, "acme.txt")
	if err := os.WriteFile(text, []byte("ACME terms"), 0o644); err != nil {
		t.Fatal(err)
	}
	node := writeLicensedPackage(t, dir, "p", `{"expression":{"operator":"AND","operands":[
		{"operator":"OR","operands":[{"kind":{"identifier":"MIT"}},{"or_later":true,"kind":{"identifier":"GPL-2.0-only"}}]},
		{"operator":"WITH","operands":[{"kind":{"identifier":"GPL-2.0-only"}}],"exception":{"identifier":"Classpath-exception-2.0"}},
		{"operator":"WITH","operands":[{"kind":{"identifier":"LicenseRef-acme","custom":true,"name":"ACME"}}],"exception":{"identifier":"acme-linking","custom":true,"name":"Linking"}}
	]},"label":"//x:license","text":"`+text+`"}`)

	doc, err := GenerateDocument(sbom.GraphConfig{}, sbom.Classifications{
		Dependencies: sbom.DependencyNodes{Direct: []sbom.NodeConfig{node}},
	}, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("GenerateDocument() error = %v", err)
	}

	var and *spdx.ConjunctiveLicenseSet
	for _, r := range relationshipsOf(doc) {
		if r.Type == spdx.RelationshipType_HasDeclaredLicense {
			and, _ = r.To[0].(*spdx.ConjunctiveLicenseSet)
		}
	}
	if and == nil || len(and.Members) != 3 {
		t.Fatalf("declared license = %#v, want an AND of 3 members", and)
	}
	if and.ID != "ConjunctiveLicenseSet-1" {
		t.Errorf("id of the AND = %q, want a stable one", and.ID)
	}
	or, ok := and.Members[0].(*spdx.DisjunctiveLicenseSet)
	if !ok || len(or.Members) != 2 {
		t.Fatalf("first member = %#v, want an OR of 2 members", and.Members[0])
	}
	if _, ok := or.Members[1].(*spdx.OrLaterOperator); !ok {
		t.Errorf("GPL-2.0-only+ = %#v, want an OrLaterOperator", or.Members[1])
	}
	listed, ok := and.Members[1].(*spdx.WithAdditionOperator)
	if !ok {
		t.Fatalf("second member = %#v, want WITH", and.Members[1])
	}
	if e, ok := listed.SubjectAddition.(*spdx.ListedLicenseException); !ok || e.ID != "https://spdx.org/licenses/Classpath-exception-2.0" {
		t.Errorf("listed exception = %#v", listed.SubjectAddition)
	}
	custom := and.Members[2].(*spdx.WithAdditionOperator)
	if a, ok := custom.SubjectAddition.(*spdx.CustomLicenseAddition); !ok || a.ID != "AdditionRef-acme-linking" || a.AdditionText != "ACME terms" {
		t.Errorf("custom exception = %#v, want AdditionRef-acme-linking with the license text", custom.SubjectAddition)
	}
}
