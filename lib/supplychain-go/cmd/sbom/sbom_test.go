package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSbomBinary_Integration(t *testing.T) {
	// Get path to sbom binary using runfiles
	r, err := runfiles.New()
	require.NoError(t, err)

	sbomBinary, err := r.Rlocation("supply-chain-go/cmd/sbom/sbom_/sbom")
	require.NoError(t, err)

	// Create temp directory
	tmpDir := t.TempDir()

	t.Run("with_root_metadata", func(t *testing.T) {
		graphPath := filepath.Join(tmpDir, "graph_with_root.json")
		outPath := filepath.Join(tmpDir, "classifications_with_root.json")

		// Write test graph with root having metadata
		graph := sbom.GraphConfig{
			SchemaVersion: "1.0",
			RootTarget:    "//test:app",
			Nodes: []sbom.NodeConfig{
				{Label: "//test:app", MetadataFile: "app.json"},
				{Label: "@@uuid//:uuid", MetadataFile: "uuid.json"},
			},
			Edges: []sbom.EdgeConfig{
				{From: "//test:app", To: "@@uuid//:uuid", Type: "depends_on"},
			},
		}
		graphBytes, _ := json.MarshalIndent(graph, "", "  ")
		require.NoError(t, os.WriteFile(graphPath, graphBytes, 0644))

		// Run sbom binary (require_root_metadata is default true)
		cmd := exec.Command(sbomBinary, "--input", graphPath, "--output", outPath)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "sbom binary failed: %s", output)

		// Read and verify output
		outBytes, err := os.ReadFile(outPath)
		require.NoError(t, err)

		var classifications sbom.Classifications
		require.NoError(t, json.Unmarshal(outBytes, &classifications))

		// Verify root component
		require.NotNil(t, classifications.RootComponent)
		assert.Equal(t, "//test:app", classifications.RootComponent.Label)

		// uuid is direct dependency
		assert.Len(t, classifications.Dependencies.Direct, 1)
		assert.Equal(t, "@@uuid//:uuid", classifications.Dependencies.Direct[0].Label)
	})

	t.Run("without_root_metadata_error", func(t *testing.T) {
		graphPath := filepath.Join(tmpDir, "graph_no_root.json")
		outPath := filepath.Join(tmpDir, "classifications_no_root.json")

		// Write test graph WITHOUT root metadata
		graph := sbom.GraphConfig{
			SchemaVersion: "1.0",
			RootTarget:    "//test:app",
			Nodes: []sbom.NodeConfig{
				{Label: "//test:lib", MetadataFile: "lib.json"},
				{Label: "@@uuid//:uuid", MetadataFile: "uuid.json"},
			},
			Edges: []sbom.EdgeConfig{
				{From: "//test:lib", To: "@@uuid//:uuid", Type: "depends_on"},
			},
		}
		graphBytes, _ := json.MarshalIndent(graph, "", "  ")
		require.NoError(t, os.WriteFile(graphPath, graphBytes, 0644))

		// Run sbom binary - should fail
		cmd := exec.Command(sbomBinary, "--input", graphPath, "--output", outPath)
		output, err := cmd.CombinedOutput()
		require.Error(t, err, "expected error when root has no metadata")
		assert.Contains(t, string(output), "//test:app")
		assert.Contains(t, string(output), "has no package metadata")
	})

	t.Run("without_root_metadata_allowed", func(t *testing.T) {
		graphPath := filepath.Join(tmpDir, "graph_no_root_allowed.json")
		outPath := filepath.Join(tmpDir, "classifications_no_root_allowed.json")

		// Write test graph WITHOUT root metadata
		graph := sbom.GraphConfig{
			SchemaVersion: "1.0",
			RootTarget:    "//test:app",
			Nodes: []sbom.NodeConfig{
				{Label: "//test:lib", MetadataFile: "lib.json"},
				{Label: "@@uuid//:uuid", MetadataFile: "uuid.json"},
			},
			Edges: []sbom.EdgeConfig{
				{From: "//test:lib", To: "@@uuid//:uuid", Type: "depends_on"},
			},
		}
		graphBytes, _ := json.MarshalIndent(graph, "", "  ")
		require.NoError(t, os.WriteFile(graphPath, graphBytes, 0644))

		// Run sbom binary with flag to allow missing root metadata
		cmd := exec.Command(sbomBinary, "--input", graphPath, "--output", outPath, "--allow_missing_root_metadata")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "sbom binary failed: %s", output)

		// Read and verify output
		outBytes, err := os.ReadFile(outPath)
		require.NoError(t, err)

		var classifications sbom.Classifications
		require.NoError(t, json.Unmarshal(outBytes, &classifications))

		// No root component when root has no metadata
		assert.Nil(t, classifications.RootComponent)

		// lib is direct (no incoming edges)
		require.Len(t, classifications.Dependencies.Direct, 1)
		assert.Equal(t, "//test:lib", classifications.Dependencies.Direct[0].Label)

		// uuid is transitive
		require.Len(t, classifications.Dependencies.Transitive, 1)
		assert.Equal(t, "@@uuid//:uuid", classifications.Dependencies.Transitive[0].Label)
	})
}
