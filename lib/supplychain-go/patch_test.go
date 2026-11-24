package supplychain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// sha256 of the empty string and the sha1 of "abc", as SRI strings.
	sriSha256 = "sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="
	sriSha1   = "sha1-qZk+NkcGgWq6PiVxeFDCbJzQ2J0="
)

func TestPatchAttributeParser(t *testing.T) {
	parse := func(s string) (*PatchAttribute, error) {
		return PatchAttributeDescriptor.Parser(strings.NewReader(s))
	}

	a, err := parse(`{"description":"Fix.","integrity":["` + sriSha256 + `"],"resolves":["CVE-2024-1"],"type":"backport","url":"https://example.com/a.patch"}`)
	require.NoError(t, err)
	assert.Equal(t, &PatchAttribute{
		Description: "Fix.",
		Integrity:   []string{sriSha256},
		Type:        PatchBackport,
		URL:         "https://example.com/a.patch",
		Resolves:    []string{"CVE-2024-1"},
	}, a)

	for name, input := range map[string]string{
		"no integrity":      `{"description":"x","integrity":[],"type":"backport","url":"u"}`,
		"bad type":          `{"description":"x","integrity":["` + sriSha256 + `"],"type":"bogus","url":"u"}`,
		"no source":         `{"description":"x","integrity":["` + sriSha256 + `"],"type":"backport"}`,
		"two sources":       `{"description":"x","integrity":["` + sriSha256 + `"],"type":"backport","url":"u","file":"f","file_exec_path":"f"}`,
		"malformed digest":  `{"description":"x","integrity":["sha256-!!!"],"type":"backport","url":"u"}`,
		"wrong size digest": `{"description":"x","integrity":["sha256-AAAA"],"type":"backport","url":"u"}`,
		"no exec path":      `{"description":"x","integrity":["` + sriSha256 + `"],"type":"backport","file":"f"}`,
	} {
		_, err := parse(input)
		assert.Error(t, err, name)
	}
}

func TestPatchAttributeChecksums(t *testing.T) {
	p := &PatchAttribute{Integrity: []string{sriSha1, sriSha256, "md5-1B2M2Y8AsgTpgAmY7PhCfg=="}}
	checksums, err := p.Checksums()
	require.NoError(t, err)
	// sha1 and md5 are not supported and dropped.
	assert.Equal(t, []Checksum{
		{Algorithm: "sha256", Hex: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	}, checksums)

	for name, integrity := range map[string]string{
		"no algorithm": "abc",
		"bad base64":   "sha256-!!!",
		"bad length":   "sha256-qZk+NkcGgWq6PiVxeFDCbJzQ2J0=",
	} {
		_, err := (&PatchAttribute{Integrity: []string{integrity}}).Checksums()
		assert.Error(t, err, name)
	}
}

func TestPatchAttributeName(t *testing.T) {
	assert.Equal(t, "a.patch", (&PatchAttribute{URL: "https://example.com/p/a.patch?raw=1"}).Name())
	assert.Equal(t, "b.patch", (&PatchAttribute{File: "pkg/patches/b.patch"}).Name())
	assert.Equal(t, "patch", (&PatchAttribute{URL: "https://example.com"}).Name())
}

func TestPatchAttributeFileComment(t *testing.T) {
	assert.Equal(t, "Patch file pkg/b.patch in the repository that declares the package.", (&PatchAttribute{File: "pkg/b.patch"}).FileComment())
	assert.Empty(t, (&PatchAttribute{URL: "https://example.com/a.patch"}).FileComment())
}

func TestGetPatchAttributes(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
		return p
	}
	first := write("first.json", `{"description":"first","integrity":["`+sriSha256+`"],"type":"unofficial","url":"https://example.com/first.patch"}`)
	second := write("second.json", `{"description":"second","integrity":["`+sriSha1+`"],"type":"monkey","file":"second.patch","file_exec_path":"/x/second.patch"}`)

	m, err := ReadPackageMetadata(strings.NewReader(`{
		"label": "//:pkg",
		"purl": "pkg:generic/pkg@1",
		"attributes": {"` + PatchAttributeKind + `": ["` + first + `", "` + second + `"]}
	}`))
	require.NoError(t, err)

	patches, err := GetPatchAttributes(m)
	require.NoError(t, err)
	require.Len(t, patches, 2)
	assert.Equal(t, "first", patches[0].Description)
	assert.Equal(t, "second", patches[1].Description)

	none, err := ReadPackageMetadata(strings.NewReader(`{"label": "//:pkg", "purl": "pkg:generic/pkg@1", "attributes": {}}`))
	require.NoError(t, err)
	patches, err = GetPatchAttributes(none)
	require.NoError(t, err)
	assert.Empty(t, patches)
}

func TestPatchAttributeWarnIfNoSupportedIntegrity(t *testing.T) {
	md5 := "md5-1B2M2Y8AsgTpgAmY7PhCfg=="
	for name, tc := range map[string]struct {
		integrity []string
		warns     bool
	}{
		"only unsupported":  {[]string{md5, sriSha1, "sha3-512-AAAA"}, true},
		"one supported":     {[]string{md5, sriSha256}, false},
		"only supported":    {[]string{sriSha256}, false},
		"malformed entries": {[]string{"abc"}, false}, // The parser rejects it; no warning on top.
	} {
		var out strings.Builder
		(&PatchAttribute{URL: "https://example.com/a.patch", Integrity: tc.integrity}).WarnIfNoSupportedIntegrity(&out, "//lib:dep")
		if tc.warns {
			assert.Contains(t, out.String(), "warning", name)
			assert.Contains(t, out.String(), "a.patch", name)
			assert.Contains(t, out.String(), "//lib:dep", name)
			assert.Contains(t, out.String(), "sha256", name)
		} else {
			assert.Empty(t, out.String(), name)
		}
	}
}
