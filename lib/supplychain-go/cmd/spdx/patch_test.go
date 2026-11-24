package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

const (
	sha256SRI = "sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="
	sha256Hex = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

const urlPatch = `{"description":"Fix overflow.","integrity":["` + sha256SRI + `","md5-1B2M2Y8AsgTpgAmY7PhCfg=="],"resolves":["CVE-2024-1"],"type":"backport","url":"https://example.com/fix.patch"}`

const filePatch = `{"description":"Local tweak.","integrity":["` + sha256SRI + `"],"type":"unofficial","file":"pkg/local.patch","file_exec_path":"pkg/local.patch"}`

const fileComment = "Patch file pkg/local.patch in the repository that declares the package."

func patchedGraph(t *testing.T, attr string) (sbom.GraphConfig, sbom.Classifications) {
	t.Helper()
	dir := t.TempDir()
	patch := filepath.Join(dir, "patch.json")
	if err := os.WriteFile(patch, []byte(attr), 0o644); err != nil {
		t.Fatal(err)
	}
	meta := filepath.Join(dir, "dep.json")
	content := `{"label":"//lib:dep","purl":"pkg:generic/dep@1","attributes":{"build.bazel.attribute.patch":["` + patch + `"]}}`
	if err := os.WriteFile(meta, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	node := sbom.NodeConfig{Label: "//lib:dep", MetadataFile: meta}
	return sbom.GraphConfig{Nodes: []sbom.NodeConfig{node}},
		sbom.Classifications{Dependencies: sbom.DependencyNodes{Direct: []sbom.NodeConfig{node}}}
}

func TestGenerateDocument23_Patches(t *testing.T) {
	graph, cls := patchedGraph(t, urlPatch)
	doc, err := GenerateDocument23(graph, cls, time.Unix(1700000000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Packages) != 2 {
		t.Fatalf("got %d packages, want 2", len(doc.Packages))
	}
	p := doc.Packages[1]
	if p.PackageName != "fix.patch" || p.PackageDescription != "Fix overflow." || p.PackageDownloadLocation != "https://example.com/fix.patch" || p.PrimaryPackagePurpose != "FILE" {
		t.Errorf("unexpected patch package %+v", p)
	}
	if len(p.PackageChecksums) != 1 || p.PackageChecksums[0].Algorithm != "SHA256" || p.PackageChecksums[0].Value != sha256Hex {
		t.Errorf("checksums = %+v, want only the sha256", p.PackageChecksums)
	}
	var found bool
	for _, r := range doc.Relationships {
		if r.Relationship == "PATCH_APPLIED" {
			found = true
			if r.RefA.ElementRefID != p.PackageSPDXIdentifier || r.RefB.ElementRefID != doc.Packages[0].PackageSPDXIdentifier {
				t.Errorf("relationship %+v does not link the patch to the package", r)
			}
		}
	}
	if !found {
		t.Error("PATCH_APPLIED relationship missing")
	}
}

func TestGenerateDocument301_Patches(t *testing.T) {
	graph, cls := patchedGraph(t, urlPatch)
	render := func() []byte {
		doc, err := GenerateDocument301(graph, cls, time.Unix(1700000000, 0).UTC())
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := writeDocument(doc, &buf); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	out := render()
	if !bytes.Equal(out, render()) {
		t.Error("output is not deterministic")
	}

	var parsed struct {
		Graph []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	var file, rel map[string]any
	for _, n := range parsed.Graph {
		switch n["type"] {
		case "software_File":
			file = n
		case "Relationship":
			if n["relationshipType"] == "patchedBy" {
				rel = n
			}
		}
	}
	if file == nil || rel == nil {
		t.Fatalf("patch file or patchedBy relationship missing in %s", out)
	}
	if file["software_primaryPurpose"] != "patch" || file["name"] != "fix.patch" || file["description"] != "Fix overflow." {
		t.Errorf("unexpected file %v", file)
	}
	hashes, _ := file["verifiedUsing"].([]any)
	if len(hashes) != 1 || hashes[0].(map[string]any)["hashValue"] != sha256Hex {
		t.Errorf("verifiedUsing = %v, want only the sha256", file["verifiedUsing"])
	}
}

func TestGenerateDocument23_FilePatch(t *testing.T) {
	graph, cls := patchedGraph(t, filePatch)
	doc, err := GenerateDocument23(graph, cls, time.Unix(1700000000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	p := doc.Packages[1]
	if p.PackageDownloadLocation != "NOASSERTION" || p.PackageComment != fileComment {
		t.Errorf("download location = %q, comment = %q", p.PackageDownloadLocation, p.PackageComment)
	}
}

func TestGenerateDocument301_FilePatch(t *testing.T) {
	graph, cls := patchedGraph(t, filePatch)
	doc, err := GenerateDocument301(graph, cls, time.Unix(1700000000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writeDocument(doc, &buf); err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Graph []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	for _, n := range parsed.Graph {
		if n["type"] == "software_File" {
			if n["comment"] != fileComment || n["externalRef"] != nil {
				t.Errorf("unexpected file %v", n)
			}
			return
		}
	}
	t.Fatal("patch file missing")
}

const unsupportedPatch = `{"description":"x","integrity":["md5-1B2M2Y8AsgTpgAmY7PhCfg=="],"type":"backport","url":"https://example.com/md5.patch"}`

func TestPatchWithoutSupportedIntegrityWarns(t *testing.T) {
	for name, tc := range map[string]struct {
		attr string
		warn bool
	}{
		"unsupported only": {unsupportedPatch, true},
		"supported":        {urlPatch, false},
	} {
		for format, generate := range map[string]func(sbom.GraphConfig, sbom.Classifications) error{
			"2.3": func(g sbom.GraphConfig, c sbom.Classifications) error {
				_, err := GenerateDocument23(g, c, time.Unix(1700000000, 0).UTC())
				return err
			},
			"3.0.1": func(g sbom.GraphConfig, c sbom.Classifications) error {
				_, err := GenerateDocument301(g, c, time.Unix(1700000000, 0).UTC())
				return err
			},
		} {
			var out bytes.Buffer
			warnings = &out
			graph, cls := patchedGraph(t, tc.attr)
			if err := generate(graph, cls); err != nil {
				t.Fatalf("%s %s: %v", name, format, err)
			}
			if got := out.String(); (got != "") != tc.warn {
				t.Errorf("%s %s: warnings = %q, want a warning: %v", name, format, got, tc.warn)
			}
		}
	}
	warnings = os.Stderr
}
