package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writePackage(t *testing.T, dir, name, licenseJSON string) sbom.NodeConfig {
	t.Helper()
	attributes := "{}"
	if licenseJSON != "" {
		licensePath := filepath.Join(dir, name+".license.json")
		require.NoError(t, os.WriteFile(licensePath, []byte(licenseJSON), 0o644))
		attributes = fmt.Sprintf(`{"build.bazel.attribute.license": %q}`, licensePath)
	}
	metadataPath := filepath.Join(dir, name+".metadata.json")
	require.NoError(t, os.WriteFile(metadataPath, []byte(fmt.Sprintf(
		`{"label": "@@//%s:package_metadata", "purl": "pkg:generic/%s@1.0.0", "attributes": %s}`,
		name, name, attributes,
	)), 0o644))
	return sbom.NodeConfig{Label: "//" + name, MetadataFile: metadataPath}
}

func TestGenerateBOMLicenses(t *testing.T) {
	dir := t.TempDir()

	none := writePackage(t, dir, "none", "")
	single := writePackage(t, dir, "single", `{"kind":{"identifier":"MIT","name":"MIT License"},"label":"//single:license"}`)
	deprecated := writePackage(t, dir, "deprecated", `{"kind":{"identifier":"GPL-2.0+","name":"GPL 2.0 or later"},"label":"//deprecated:license"}`)
	custom := writePackage(t, dir, "custom", `{"kind":{"identifier":"LicenseRef-acme","name":"ACME License"},"label":"//custom:license"}`)
	dual := writePackage(t, dir, "dual", `{"expression":{"operand":"OR","operands":[
		{"kind":{"identifier":"Apache-2.0","name":"Apache"}},
		{"kind":{"identifier":"MIT","name":"MIT License"}}]},"label":"//dual:license"}`)
	exception := writePackage(t, dir, "exception", `{"expression":{"operand":"OR","operands":[
		{"kind":{"identifier":"LicenseRef-acme","name":"ACME License"}},
		{"operand":"WITH","operands":[{"kind":{"identifier":"GPL-2.0-only","name":"GPL"}}],
		 "exception":{"identifier":"Classpath-exception-2.0","name":"Classpath"}}]},"label":"//exception:license"}`)

	bom, err := GenerateBOM(
		sbom.GraphConfig{},
		sbom.Classifications{
			Dependencies: sbom.DependencyNodes{
				Direct: []sbom.NodeConfig{none, single, deprecated, custom, dual, exception},
			},
		},
	)
	require.NoError(t, err)

	licenses := map[string]*cdx.Licenses{}
	for _, c := range *bom.Components {
		licenses[c.Name] = c.Licenses
	}

	assert.Nil(t, licenses["none"])
	assert.Equal(t, &cdx.Licenses{{License: &cdx.License{ID: "MIT"}}}, licenses["single"])
	assert.Equal(t, &cdx.Licenses{{License: &cdx.License{ID: "GPL-2.0+"}}}, licenses["deprecated"])
	assert.Equal(t, &cdx.Licenses{{License: &cdx.License{Name: "ACME License"}}}, licenses["custom"])
	assert.Equal(t, &cdx.Licenses{{Expression: "Apache-2.0 OR MIT"}}, licenses["dual"])
	assert.Equal(t,
		&cdx.Licenses{{Expression: "LicenseRef-acme OR GPL-2.0-only WITH Classpath-exception-2.0"}},
		licenses["exception"],
	)
}
