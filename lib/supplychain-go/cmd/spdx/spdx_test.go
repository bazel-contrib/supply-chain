package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

func TestGenerateDocument_RetainsPackageEdges(t *testing.T) {
	for _, wrappers := range []bool{false, true} {
		t.Run(fmt.Sprintf("wrappers=%t", wrappers), func(t *testing.T) {
			dir := t.TempDir()
			root := sbom.NodeConfig{Label: "//app:root", MetadataFile: writePackageMetadata(t, dir, "root", "pkg:golang/example.com/app@v1")}
			direct := sbom.NodeConfig{Label: "//module_a:library", MetadataFile: writePackageMetadata(t, dir, "direct", "pkg:golang/example.com/module_a@v1")}
			dep := sbom.NodeConfig{Label: "//module_b:other", MetadataFile: writePackageMetadata(t, dir, "dep", "pkg:golang/example.com/module_b@v1")}
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
			doc, err := GenerateDocument(graph, classifications)
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

func writePackageMetadata(t *testing.T, dir, name, purl string) string {
	t.Helper()
	path := filepath.Join(dir, name+".json")
	if err := os.WriteFile(path, []byte(fmt.Sprintf(`{"label":"//fixture:%s","purl":%q,"attributes":{}}`, name, purl)), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
