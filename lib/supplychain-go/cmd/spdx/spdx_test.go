package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// writePackage writes a package_metadata file, optionally with a license
// attribute, and returns the node for it.
func writePackage(t *testing.T, dir, name, licenseJSON string) sbom.NodeConfig {
	t.Helper()
	attributes := "{}"
	if licenseJSON != "" {
		licensePath := filepath.Join(dir, name+".license.json")
		writeFile(t, licensePath, licenseJSON)
		attributes = fmt.Sprintf(`{"build.bazel.attribute.license": %q}`, licensePath)
	}
	metadataPath := filepath.Join(dir, name+".metadata.json")
	writeFile(t, metadataPath, fmt.Sprintf(
		`{"label": "@@//%s:package_metadata", "purl": "pkg:generic/%s@1.0.0", "attributes": %s}`,
		name, name, attributes,
	))
	return sbom.NodeConfig{Label: "//" + name, MetadataFile: metadataPath}
}

func TestGenerateDocumentLicenses(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "LICENSE")
	writeFile(t, textPath, "ACME license text")

	root := writePackage(t, dir, "root", "")
	single := writePackage(t, dir, "single", `{"kind":{"identifier":"MIT","name":"MIT License"},"label":"//single:license"}`)
	custom := writePackage(t, dir, "custom", fmt.Sprintf(
		`{"kind":{"identifier":"LicenseRef-acme","name":"ACME License"},"label":"//custom:license","text":%q}`, textPath))
	dual := writePackage(t, dir, "dual", `{"expression":{"operator":"OR","operands":[
		{"kind":{"identifier":"MIT","name":"MIT License"}},
		{"kind":{"identifier":"LicenseRef-acme","name":"ACME License"}}]},"label":"//dual:license"}`)
	exception := writePackage(t, dir, "exception", `{"expression":{"operator":"WITH","operands":[
		{"kind":{"identifier":"GPL-2.0-only","name":"GPL"}}],
		"exception":{"identifier":"AdditionRef-link","name":"Linking"}},"label":"//exception:license"}`)

	doc, err := GenerateDocument(
		sbom.GraphConfig{},
		sbom.Classifications{
			RootComponent: &root,
			Dependencies: sbom.DependencyNodes{
				Direct: []sbom.NodeConfig{single, custom, dual, exception},
			},
		},
	)
	require.NoError(t, err)

	declared := map[string]string{}
	for _, p := range doc.Packages {
		declared[p.PackageName] = p.PackageLicenseDeclared
	}
	assert.Equal(t, map[string]string{
		"root":      "NOASSERTION",
		"single":    "MIT",
		"custom":    "LicenseRef-acme",
		"dual":      "MIT OR LicenseRef-acme",
		"exception": "GPL-2.0-only WITH AdditionRef-link",
	}, declared)

	// The custom license is defined once, with its text, and the custom
	// exception is not defined as a license.
	require.Len(t, doc.OtherLicenses, 1)
	assert.Equal(t, "LicenseRef-acme", doc.OtherLicenses[0].LicenseIdentifier)
	assert.Equal(t, "ACME License", doc.OtherLicenses[0].LicenseName)
	assert.Equal(t, "ACME license text", doc.OtherLicenses[0].ExtractedText)
}
