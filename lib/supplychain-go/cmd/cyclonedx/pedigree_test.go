package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
)

const sha256SRI = "sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="

func generatePatched(t *testing.T, dir string, attrs ...string) (cdx.Component, error) {
	t.Helper()
	list := ""
	for i, a := range attrs {
		p := filepath.Join(dir, "patch"+strconv.Itoa(i)+".json")
		if err := os.WriteFile(p, []byte(a), 0o644); err != nil {
			t.Fatal(err)
		}
		if i > 0 {
			list += ","
		}
		list += `"` + p + `"`
	}
	meta := filepath.Join(dir, "dep.json")
	content := `{"label":"//lib:dep","purl":"pkg:generic/dep@1","attributes":{"build.bazel.attribute.patch":[` + list + `]}}`
	if len(attrs) == 0 {
		content = `{"label":"//lib:dep","purl":"pkg:generic/dep@1","attributes":{}}`
	}
	if err := os.WriteFile(meta, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	node := sbom.NodeConfig{Label: "//lib:dep", MetadataFile: meta}
	bom, err := GenerateBOM(sbom.GraphConfig{Nodes: []sbom.NodeConfig{node}},
		sbom.Classifications{Dependencies: sbom.DependencyNodes{Direct: []sbom.NodeConfig{node}}})
	if err != nil {
		return cdx.Component{}, err
	}
	return (*bom.Components)[0], nil
}

func TestGenerateBOM_Patches(t *testing.T) {
	dir := t.TempDir()
	diff := "--- a/x\n+++ b/x\n"
	patchFile := filepath.Join(dir, "b.patch")
	if err := os.WriteFile(patchFile, []byte(diff), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := generatePatched(t, dir,
		`{"description":"Fix overflow.","integrity":["`+sha256SRI+`"],"resolves":["CVE-2024-1","https://example.com/issues/7"],"type":"backport","url":"https://example.com/a.patch"}`,
		`{"description":"Local tweak.","integrity":["`+sha256SRI+`"],"type":"unofficial","file":"pkg/b.patch","file_exec_path":"`+patchFile+`"}`)
	if err != nil {
		t.Fatalf("GenerateBOM() error = %v", err)
	}

	if c.Pedigree == nil || c.Pedigree.Patches == nil {
		t.Fatal("pedigree patches missing")
	}
	want := []cdx.Patch{
		{
			Type: cdx.PatchTypeBackport,
			Diff: &cdx.Diff{URL: "https://example.com/a.patch"},
			Resolves: &[]cdx.Issue{
				{ID: "CVE-2024-1", Type: cdx.IssueTypeSecurity},
				{ID: "https://example.com/issues/7", Type: cdx.IssueTypeDefect, References: &[]string{"https://example.com/issues/7"}},
			},
		},
		{
			Type: cdx.PatchTypeUnofficial,
			Diff: &cdx.Diff{Text: &cdx.AttachedText{
				Content:     base64.StdEncoding.EncodeToString([]byte(diff)),
				ContentType: "text/x-patch",
				Encoding:    "base64",
			}},
		},
	}
	if !reflect.DeepEqual(*c.Pedigree.Patches, want) {
		t.Errorf("patches = %+v, want %+v", *c.Pedigree.Patches, want)
	}
	wantNotes := "a.patch: Fix overflow.\nb.patch: Local tweak. Patch file pkg/b.patch in the repository that declares the package."
	if c.Pedigree.Notes != wantNotes {
		t.Errorf("notes = %q, want %q", c.Pedigree.Notes, wantNotes)
	}
}

func TestGenerateBOM_MissingPatchFile(t *testing.T) {
	dir := t.TempDir()
	_, err := generatePatched(t, dir, `{"description":"x","integrity":["`+sha256SRI+`"],"type":"unofficial","file":"b.patch","file_exec_path":"`+filepath.Join(dir, "missing.patch")+`"}`)
	if err == nil {
		t.Fatal("GenerateBOM() succeeded, want an error for the missing patch file")
	}
}

func TestGenerateBOM_NoPatchesNoPedigree(t *testing.T) {
	c, err := generatePatched(t, t.TempDir())
	if err != nil {
		t.Fatalf("GenerateBOM() error = %v", err)
	}
	if c.Pedigree != nil {
		t.Errorf("pedigree = %+v, want nil", c.Pedigree)
	}
}
