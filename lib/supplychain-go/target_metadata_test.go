package supplychain

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/label"
)

func TestReadTargetMetadataValid(t *testing.T) {
	type ExpectedDependency struct {
		Scope dependencyScope
		Path  string
	}

	type ExpectedData struct {
		Label              string
		PackageMetadata    []string
		ExpectedDependency []ExpectedDependency
	}

	cases := []struct {
		Name     string
		Input    string
		Expected ExpectedData
	}{
		{
			Name: "minimal",
			Input: `
			{
				"label": "@workspace//foo/bar:baz",
				"package_metadata": [],
				"dependencies": []
			}
			`,
			Expected: ExpectedData{
				Label:              "@workspace//foo/bar:baz",
				PackageMetadata:    []string{},
				ExpectedDependency: []ExpectedDependency{},
			},
		},
		{
			Name: "package-metadata-paths",
			Input: `
			{
				"label": "@workspace//foo/bar:baz",
				"package_metadata": [
					"path/to/a.metadata.json",
					"path/to/b.metadata.json"
				],
				"dependencies": []
			}
			`,
			Expected: ExpectedData{
				Label: "@workspace//foo/bar:baz",
				PackageMetadata: []string{
					"path/to/a.metadata.json",
					"path/to/b.metadata.json",
				},
				ExpectedDependency: []ExpectedDependency{},
			},
		},
		{
			Name: "dependencies",
			Input: `
			{
				"label": "@workspace//foo/bar:baz",
				"package_metadata": [],
				"dependencies": [
					{"scope": "tool", "path": "path/to/tool.metadata.json"},
					{"scope": "runtime", "path": "path/to/runtime.metadata.json"},
					{"scope": "dynamic", "path": "path/to/dynamic.metadata.json"},
					{"scope": "bundled", "path": "path/to/bundled.metadata.json"}
				]
			}
			`,
			Expected: ExpectedData{
				Label:           "@workspace//foo/bar:baz",
				PackageMetadata: []string{},
				ExpectedDependency: []ExpectedDependency{
					{Scope: DependencyScopeTool, Path: "path/to/tool.metadata.json"},
					{Scope: DependencyScopeRuntime, Path: "path/to/runtime.metadata.json"},
					{Scope: DependencyScopeDynamic, Path: "path/to/dynamic.metadata.json"},
					{Scope: DependencyScopeBundled, Path: "path/to/bundled.metadata.json"},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(
			fmt.Sprintf("Deserialize %s", c.Name),
			func(t *testing.T) {
				m, err := ReadTargetMetadata(strings.NewReader(c.Input))
				assert.Nil(t, err)
				assert.Equal(t, label.MustParse(c.Expected.Label), m.GetLabel())

				raw := m.(*targetMetadata)
				assert.Equal(t, c.Expected.PackageMetadata, raw.PackageMetadata)

				assert.Equal(t, len(c.Expected.ExpectedDependency), len(raw.Dependencies))
				for i, expected := range c.Expected.ExpectedDependency {
					assert.Equal(t, string(expected.Scope), raw.Dependencies[i].Scope)
					assert.Equal(t, expected.Path, raw.Dependencies[i].Path)
				}
			})
		t.Run(
			fmt.Sprintf("Serialize %s", c.Name),
			func(t *testing.T) {
				m, err := ReadTargetMetadata(strings.NewReader(c.Input))
				assert.Nil(t, err)

				b, err := json.Marshal(m)
				assert.Nil(t, err)
				assert.JSONEq(t, c.Input, string(b))
			})
	}
}

func TestReadTargetMetadataInvalid(t *testing.T) {
	_, err := ReadTargetMetadata(strings.NewReader(`not json`))
	assert.NotNil(t, err)
}

func TestReadWriteTargetMetadataFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "target.metadata.json")

	original := &targetMetadata{
		Label:           label.MustParse("@workspace//foo/bar:baz"),
		PackageMetadata: []string{"some/package_metadata.json"},
		Dependencies: []rawDependency{
			{Scope: string(DependencyScopeTool), Path: "some/tool.metadata.json"},
		},
	}

	assert.Nil(t, WriteTargetMetadataToFile(path, original))

	roundTripped, err := ReadTargetMetadataFromFile(path)
	assert.Nil(t, err)
	assert.Equal(t, original.GetLabel(), roundTripped.GetLabel())

	raw := roundTripped.(*targetMetadata)
	assert.Equal(t, original.PackageMetadata, raw.PackageMetadata)
	assert.Equal(t, original.Dependencies, raw.Dependencies)
}

func TestReadTargetMetadataFromFileMissing(t *testing.T) {
	_, err := ReadTargetMetadataFromFile(filepath.Join(t.TempDir(), "does-not-exist.json"))
	assert.NotNil(t, err)
}

func TestGetDependencies(t *testing.T) {
	dir := t.TempDir()

	toolPath := filepath.Join(dir, "tool.metadata.json")
	assert.Nil(t, WriteTargetMetadataToFile(toolPath, &targetMetadata{
		Label: label.MustParse("@workspace//tool:compiler"),
	}))

	runtimePath := filepath.Join(dir, "runtime.metadata.json")
	assert.Nil(t, WriteTargetMetadataToFile(runtimePath, &targetMetadata{
		Label: label.MustParse("@workspace//lib:runtime"),
	}))

	parent := &targetMetadata{
		Label: label.MustParse("@workspace//foo:bar"),
		Dependencies: []rawDependency{
			{Scope: string(DependencyScopeTool), Path: toolPath},
			{Scope: string(DependencyScopeRuntime), Path: runtimePath},
		},
	}

	deps, err := parent.GetDependencies()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(deps))

	assert.Equal(t, DependencyScopeTool, deps[0].GetScope())
	assert.Equal(t, label.MustParse("@workspace//tool:compiler"), deps[0].GetTargetMetadata().GetLabel())

	assert.Equal(t, DependencyScopeRuntime, deps[1].GetScope())
	assert.Equal(t, label.MustParse("@workspace//lib:runtime"), deps[1].GetTargetMetadata().GetLabel())
}

func TestGetDependenciesMissingFile(t *testing.T) {
	parent := &targetMetadata{
		Label: label.MustParse("@workspace//foo:bar"),
		Dependencies: []rawDependency{
			{Scope: string(DependencyScopeTool), Path: filepath.Join(t.TempDir(), "missing.json")},
		},
	}

	_, err := parent.GetDependencies()
	assert.NotNil(t, err)
}

func TestGetPackageMetadata(t *testing.T) {
	dir := t.TempDir()

	pmPath := filepath.Join(dir, "pm.metadata.json")
	assert.Nil(t, WritePackageMetadataToFile(pmPath, &packageMetadata{
		Label: label.MustParse("@package_metadata//foo/bar"),
		PURL: *packageurl.NewPackageURL(
			"github",
			"bazel-contrib",
			"supply-chain",
			"",
			packageurl.Qualifiers{},
			"",
		),
		Attributes: map[string]string{},
	}))

	parent := &targetMetadata{
		Label:           label.MustParse("@workspace//foo:bar"),
		PackageMetadata: []string{pmPath},
	}

	pm, err := parent.GetPackageMetadata()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(pm))
	assert.Equal(t, label.MustParse("@package_metadata//foo/bar"), pm[0].GetLabel())
}

func TestGetPackageMetadataMissingFile(t *testing.T) {
	parent := &targetMetadata{
		Label:           label.MustParse("@workspace//foo:bar"),
		PackageMetadata: []string{filepath.Join(t.TempDir(), "missing.json")},
	}

	_, err := parent.GetPackageMetadata()
	assert.NotNil(t, err)
}
