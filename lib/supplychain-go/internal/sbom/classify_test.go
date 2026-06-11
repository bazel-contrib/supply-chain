package sbom_test

import (
	"testing"

	"github.com/bazel-contrib/supply-chain/lib/supplychain-go/internal/sbom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeClassifications_WithRootMetadata(t *testing.T) {
	graph := sbom.GraphConfig{
		SchemaVersion: "1.0",
		RootTarget:    "//app:binary",
		Nodes: []sbom.NodeConfig{
			{Label: "//app:binary", MetadataFile: "app.json"},
			{Label: "//lib:lib", MetadataFile: "lib.json"},
			{Label: "@@uuid//:uuid", MetadataFile: "uuid.json"},
		},
		Edges: []sbom.EdgeConfig{
			{From: "//app:binary", To: "//lib:lib", Type: "depends_on"},
			{From: "//lib:lib", To: "@@uuid//:uuid", Type: "depends_on"},
		},
	}

	classifications, err := sbom.ComputeClassifications(graph, false)
	require.NoError(t, err)

	// Root has metadata, should be root component
	require.NotNil(t, classifications.RootComponent)
	assert.Equal(t, "//app:binary", classifications.RootComponent.Label)

	// lib is direct dependency of root
	require.Len(t, classifications.Dependencies.Direct, 1)
	assert.Equal(t, "//lib:lib", classifications.Dependencies.Direct[0].Label)

	// uuid is transitive
	require.Len(t, classifications.Dependencies.Transitive, 1)
	assert.Equal(t, "@@uuid//:uuid", classifications.Dependencies.Transitive[0].Label)
}

func TestComputeClassifications_BinaryEmbeddingLibrary_RequireMetadata(t *testing.T) {
	// Binary has no metadata - should error when metadata is required
	graph := sbom.GraphConfig{
		SchemaVersion: "1.0",
		RootTarget:    "//app:binary",
		Nodes: []sbom.NodeConfig{
			{Label: "//app:lib", MetadataFile: "lib.json"},
			{Label: "@@uuid//:uuid", MetadataFile: "uuid.json"},
		},
		Edges: []sbom.EdgeConfig{
			{From: "//app:lib", To: "@@uuid//:uuid", Type: "depends_on"},
		},
	}

	_, err := sbom.ComputeClassifications(graph, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "//app:binary")
	assert.Contains(t, err.Error(), "has no package metadata")
}

func TestComputeClassifications_BinaryEmbeddingLibrary_AllowMissing(t *testing.T) {
	// Binary has no metadata - allowed when flag is set
	graph := sbom.GraphConfig{
		SchemaVersion: "1.0",
		RootTarget:    "//app:binary",
		Nodes: []sbom.NodeConfig{
			{Label: "//app:lib", MetadataFile: "lib.json"},
			{Label: "@@uuid//:uuid", MetadataFile: "uuid.json"},
		},
		Edges: []sbom.EdgeConfig{
			{From: "//app:lib", To: "@@uuid//:uuid", Type: "depends_on"},
		},
	}

	classifications, err := sbom.ComputeClassifications(graph, true)
	require.NoError(t, err)

	// No root component when root has no metadata
	assert.Nil(t, classifications.RootComponent)

	// lib is direct (no incoming edges)
	require.Len(t, classifications.Dependencies.Direct, 1)
	assert.Equal(t, "//app:lib", classifications.Dependencies.Direct[0].Label)

	// uuid is transitive
	require.Len(t, classifications.Dependencies.Transitive, 1)
	assert.Equal(t, "@@uuid//:uuid", classifications.Dependencies.Transitive[0].Label)
}

func TestComputeClassifications_NoRootMetadata_MultipleNodesWithoutIncoming(t *testing.T) {
	// Multiple nodes without incoming edges, root has no metadata
	graph := sbom.GraphConfig{
		SchemaVersion: "1.0",
		RootTarget:    "//app:binary",
		Nodes: []sbom.NodeConfig{
			{Label: "//lib1:lib1", MetadataFile: "lib1.json"},
			{Label: "//lib2:lib2", MetadataFile: "lib2.json"},
			{Label: "@@uuid//:uuid", MetadataFile: "uuid.json"},
		},
		Edges: []sbom.EdgeConfig{
			{From: "//lib1:lib1", To: "@@uuid//:uuid", Type: "depends_on"},
		},
	}

	classifications, err := sbom.ComputeClassifications(graph, true)
	require.NoError(t, err)

	// No root component when root has no metadata
	assert.Nil(t, classifications.RootComponent)

	// Both lib1 and lib2 have no incoming edges, so both are direct
	require.Len(t, classifications.Dependencies.Direct, 2)
	directLabels := []string{classifications.Dependencies.Direct[0].Label, classifications.Dependencies.Direct[1].Label}
	assert.Contains(t, directLabels, "//lib1:lib1")
	assert.Contains(t, directLabels, "//lib2:lib2")

	// uuid is transitive
	require.Len(t, classifications.Dependencies.Transitive, 1)
	assert.Equal(t, "@@uuid//:uuid", classifications.Dependencies.Transitive[0].Label)
}

func TestComputeClassifications_EmptyGraph(t *testing.T) {
	graph := sbom.GraphConfig{
		SchemaVersion: "1.0",
		RootTarget:    "",
		Nodes:         []sbom.NodeConfig{},
		Edges:         []sbom.EdgeConfig{},
	}

	classifications, err := sbom.ComputeClassifications(graph, false)
	require.NoError(t, err)

	assert.Nil(t, classifications.RootComponent)
	assert.Empty(t, classifications.Dependencies.Direct)
	assert.Empty(t, classifications.Dependencies.Transitive)
}

func TestComputeClassifications_SingleNodeNoRoot(t *testing.T) {
	// Root target has no metadata, single library node exists
	graph := sbom.GraphConfig{
		SchemaVersion: "1.0",
		RootTarget:    "//app:binary",
		Nodes: []sbom.NodeConfig{
			{Label: "//lib:lib", MetadataFile: "lib.json"},
		},
		Edges: []sbom.EdgeConfig{},
	}

	classifications, err := sbom.ComputeClassifications(graph, true)
	require.NoError(t, err)

	// No root component when root has no metadata
	assert.Nil(t, classifications.RootComponent)

	// Single library node is direct
	require.Len(t, classifications.Dependencies.Direct, 1)
	assert.Equal(t, "//lib:lib", classifications.Dependencies.Direct[0].Label)

	assert.Empty(t, classifications.Dependencies.Transitive)
}

func TestComputeClassifications_RootTargetNotInNodes(t *testing.T) {
	// Root target exists but has no metadata (not in nodes)
	graph := sbom.GraphConfig{
		SchemaVersion: "1.0",
		RootTarget:    "//app:binary",
		Nodes: []sbom.NodeConfig{
			{Label: "//lib:lib", MetadataFile: "lib.json"},
			{Label: "//dep:dep", MetadataFile: "dep.json"},
		},
		Edges: []sbom.EdgeConfig{
			{From: "//lib:lib", To: "//dep:dep", Type: "depends_on"},
		},
	}

	classifications, err := sbom.ComputeClassifications(graph, true)
	require.NoError(t, err)

	// No root component when root has no metadata
	assert.Nil(t, classifications.RootComponent)

	// lib has no incoming edges, so it's direct
	require.Len(t, classifications.Dependencies.Direct, 1)
	assert.Equal(t, "//lib:lib", classifications.Dependencies.Direct[0].Label)

	// dep is transitive
	require.Len(t, classifications.Dependencies.Transitive, 1)
	assert.Equal(t, "//dep:dep", classifications.Dependencies.Transitive[0].Label)
}

func TestComputeClassifications_MalformedGraph_RootWithMetadataNoEdgesButOtherNodes(t *testing.T) {
	// Root has metadata but no edges, yet other nodes exist - malformed graph
	graph := sbom.GraphConfig{
		SchemaVersion: "1.0",
		RootTarget:    "//app:binary",
		Nodes: []sbom.NodeConfig{
			{Label: "//app:binary", MetadataFile: "app.json"}, // Root has metadata
			{Label: "//lib:lib", MetadataFile: "lib.json"},
			{Label: "@@uuid//:uuid", MetadataFile: "uuid.json"},
		},
		Edges: []sbom.EdgeConfig{
			// Root has no edges, but lib->uuid edge exists
			{From: "//lib:lib", To: "@@uuid//:uuid", Type: "depends_on"},
		},
	}

	_, err := sbom.ComputeClassifications(graph, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "malformed graph")
	assert.Contains(t, err.Error(), "//app:binary")
	assert.Contains(t, err.Error(), "no dependency edges")
}
