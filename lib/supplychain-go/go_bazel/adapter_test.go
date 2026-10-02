package go_bazel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/coverage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRequestMapsGoExternalRepositories(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	metadata := filepath.Join(tmp, "foo.json")
	require.NoError(t, os.WriteFile(metadata, []byte(`{
  "label": "@@gazelle++go_deps+example.com_foo//:metadata",
  "purl": "pkg:golang/example.com/foo@v1.2.3"
}`), 0o644))

	requests, err := BuildRequest(coverage.GraphConfig{Nodes: []coverage.NodeConfig{{
		Label:        "@@gazelle++go_deps+example.com_foo//:metadata",
		MetadataFile: metadata,
	}}}, "/output-base")
	require.NoError(t, err)
	require.Len(t, requests, 1)
	assert.Equal(t, "pkg:golang/example.com/foo@v1.2.3", requests[0].PURL)
	assert.Equal(t, filepath.Join("/output-base", "external", "gazelle++go_deps+example.com_foo"), requests[0].Root)
}

func TestBuildRequestSkipsNonGoAndLocalPackagesAndDeduplicates(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	goMetadata := filepath.Join(tmp, "go.json")
	nonGoMetadata := filepath.Join(tmp, "non-go.json")
	for path, content := range map[string]string{
		goMetadata:    `{"label":"@repo//:metadata","purl":"pkg:golang/example.com/foo@v1"}`,
		nonGoMetadata: `{"label":"@repo//:metadata","purl":"pkg:cargo/example@v1"}`,
	} {
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	graph := coverage.GraphConfig{Nodes: []coverage.NodeConfig{
		{Label: "//local:metadata", MetadataFile: goMetadata},
		{Label: "@repo//:metadata", MetadataFile: goMetadata},
		{Label: "@repo//:metadata", MetadataFile: nonGoMetadata},
	}}

	requests, err := BuildRequest(graph, "/output-base")
	require.NoError(t, err)
	require.Len(t, requests, 1)
	assert.Equal(t, "pkg:golang/example.com/foo@v1", requests[0].PURL)
}
