package sbom

import "fmt"

type GraphConfig struct {
	SchemaVersion string       `json:"schema_version"`
	RootTarget    string       `json:"root_target"`
	Nodes         []NodeConfig `json:"nodes"`
	Edges         []EdgeConfig `json:"edges"`
}

type Classifications struct {
	RootComponent *NodeConfig     `json:"root_component,omitempty"`
	Dependencies  DependencyNodes `json:"dependencies"`
}

type DependencyNodes struct {
	Direct     []NodeConfig `json:"direct"`
	Transitive []NodeConfig `json:"transitive"`
}

func ComputeClassifications(graph GraphConfig, allowMissingRootMetadata bool) (Classifications, error) {
	scopes, err := calculateScopes(graph.RootTarget, graph.Nodes, graph.Edges, allowMissingRootMetadata)
	if err != nil {
		return Classifications{}, err
	}

	return Classifications{
		RootComponent: scopes.root,
		Dependencies: DependencyNodes{
			Direct:     scopes.direct,
			Transitive: scopes.transitive,
		},
	}, nil
}

type scopeResult struct {
	root       *NodeConfig
	direct     []NodeConfig
	transitive []NodeConfig
}

// calculateScopes determines root/direct/transitive classification
func calculateScopes(rootTarget string, nodes []NodeConfig, edges []EdgeConfig, allowMissingRootMetadata bool) (scopeResult, error) {
	result := scopeResult{
		root:       nil,
		direct:     []NodeConfig{},
		transitive: []NodeConfig{},
	}

	if rootTarget == "" {
		// No root target, treat all as direct
		for _, node := range nodes {
			result.direct = append(result.direct, node)
		}
		return result, nil
	}

	// Build adjacency list from edges
	adjacency := make(map[string][]string)
	for _, edge := range edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}

	// Check if root_target has metadata
	rootHasMetadata := false
	for _, node := range nodes {
		if node.Label == rootTarget && node.MetadataFile != "" {
			nodeCopy := node
			result.root = &nodeCopy
			rootHasMetadata = true
			break
		}
	}

	// If root has no metadata and we require it, error
	if !rootHasMetadata && !allowMissingRootMetadata {
		return result, fmt.Errorf(
			"target %s has no package metadata. "+
				"Add package_metadata to the target's BUILD file or set require_root_metadata = False on the sbom() rule",
			rootTarget,
		)
	}

	// Find direct dependencies
	directDeps := make(map[string]bool)
	if rootHasMetadata {
		// Root has metadata - normal classification
		if len(adjacency[rootTarget]) > 0 {
			// Root has edges - those are direct deps
			for _, child := range adjacency[rootTarget] {
				directDeps[child] = true
			}
		} else {
			// Root has no edges
			if len(nodes) > 1 {
				// Other nodes exist but root has no edges to them - malformed graph
				return result, fmt.Errorf(
					"malformed graph: target %s has metadata but no dependency edges, yet graph contains %d nodes. "+
						"This suggests the graph was incorrectly constructed",
					rootTarget, len(nodes),
				)
			}
			// else: only root node exists, no dependencies - valid
		}
	} else {
		// Root has no metadata (allowMissingRootMetadata=true)
		// Use heuristic: nodes without incoming edges are treated as direct
		hasIncoming := make(map[string]bool)
		for _, edge := range edges {
			hasIncoming[edge.To] = true
		}

		for _, node := range nodes {
			if node.Label != rootTarget && !hasIncoming[node.Label] {
				directDeps[node.Label] = true
			}
		}
	}

	// Classify all nodes
	for _, node := range nodes {
		label := node.Label
		if result.root != nil && label == result.root.Label {
			// Skip the root component itself
			continue
		}
		if label == rootTarget {
			// Skip root target if it has no metadata
			continue
		}

		if directDeps[label] {
			result.direct = append(result.direct, node)
		} else {
			result.transitive = append(result.transitive, node)
		}
	}

	return result, nil
}
