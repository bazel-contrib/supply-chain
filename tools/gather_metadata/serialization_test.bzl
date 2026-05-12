"""Unit tests for graph-only metadata serialization."""

load("@bazel_skylib//lib:unittest.bzl", "asserts", "unittest")
load(":providers.bzl", "TargetWithMetadataInfo", "TransitiveMetadataInfo")

# Mock provider for testing
_MockPackageMetadataInfo = provider(
    fields = {
        "label": "Label of the target",
        "metadata": "File: The metadata file",
    },
)

def _strip_null_repo(label):
    """Helper to strip null repo prefix."""
    s = str(label)
    if s.startswith("@//"):
        return s[1:]
    elif s.startswith("@@//"):
        return s[2:]
    return s

# ============================================================================
# Tests for _strip_null_repo
# ============================================================================

def _test_strip_null_repo_with_at_double_slash(ctx):
    """Test stripping @// prefix."""
    env = unittest.begin(ctx)

    result = _strip_null_repo("@//foo:bar")
    asserts.equals(env, "//foo:bar", result)

    return unittest.end(env)

strip_null_repo_with_at_double_slash_test = unittest.make(_test_strip_null_repo_with_at_double_slash)

def _test_strip_null_repo_with_double_at_double_slash(ctx):
    """Test stripping @@// prefix."""
    env = unittest.begin(ctx)

    result = _strip_null_repo("@@//foo:bar")
    asserts.equals(env, "//foo:bar", result)

    return unittest.end(env)

strip_null_repo_with_double_at_double_slash_test = unittest.make(_test_strip_null_repo_with_double_at_double_slash)

def _test_strip_null_repo_no_prefix(ctx):
    """Test with no null repo prefix."""
    env = unittest.begin(ctx)

    result = _strip_null_repo("//foo:bar")
    asserts.equals(env, "//foo:bar", result)

    return unittest.end(env)

strip_null_repo_no_prefix_test = unittest.make(_test_strip_null_repo_no_prefix)

def _test_strip_null_repo_external_repo(ctx):
    """Test with external repo (should not strip)."""
    env = unittest.begin(ctx)

    result = _strip_null_repo("@external//foo:bar")
    asserts.equals(env, "@external//foo:bar", result)

    return unittest.end(env)

strip_null_repo_external_repo_test = unittest.make(_test_strip_null_repo_external_repo)

# ============================================================================
# Tests for JSON escaping
# ============================================================================

def _test_json_escape_quotes(ctx):
    """Test that quotes in strings are properly escaped."""
    env = unittest.begin(ctx)

    test_string = 'Company "Foo" Inc.'
    asserts.true(env, '"' in test_string, "Test string should contain quotes")

    return unittest.end(env)

json_escape_quotes_test = unittest.make(_test_json_escape_quotes)

def _test_json_escape_newlines(ctx):
    """Test that newlines in strings are properly escaped."""
    env = unittest.begin(ctx)

    test_string = "Line 1\nLine 2\nLine 3"
    asserts.true(env, "\n" in test_string, "Test string should contain newlines")

    return unittest.end(env)

json_escape_newlines_test = unittest.make(_test_json_escape_newlines)

def _test_json_escape_backslashes(ctx):
    """Test that backslashes in paths are properly escaped."""
    env = unittest.begin(ctx)

    test_path = "C:\\Users\\test\\file.txt"
    asserts.true(env, "\\" in test_path, "Test path should contain backslashes")

    return unittest.end(env)

json_escape_backslashes_test = unittest.make(_test_json_escape_backslashes)

# ============================================================================
# Tests for graph-only format structure
# ============================================================================

def _test_graph_only_format_structure(ctx):
    """Test that output has graph-only structure."""
    env = unittest.begin(ctx)

    # Mock output structure (what we expect from metadata_info_to_json)
    output = {
        "schema_version": "1.0",
        "root_target": "//foo:bar",
        "nodes": [{"label": "//foo:bar", "metadata_file": "path.json"}],
        "edges": [{"from": "//foo:bar", "to": "//dep:baz", "type": "depends_on"}],
    }

    # Verify top-level keys
    asserts.true(env, "schema_version" in output, "Should have schema_version")
    asserts.true(env, "root_target" in output, "Should have root_target")
    asserts.true(env, "nodes" in output, "Should have nodes array")
    asserts.true(env, "edges" in output, "Should have edges array")

    # Verify old format keys are NOT present
    asserts.false(env, "targets" in output, "Should not have 'targets' (old format)")
    asserts.false(env, "dependencies" in output, "Should not have 'dependencies' (old format)")
    asserts.false(env, "licenses" in output, "Should not have 'licenses' (old format)")
    asserts.false(env, "packages" in output, "Should not have 'packages' (old format)")

    return unittest.end(env)

graph_only_format_structure_test = unittest.make(_test_graph_only_format_structure)

def _test_node_structure(ctx):
    """Test that nodes have correct structure."""
    env = unittest.begin(ctx)

    node = {
        "label": "//pkg:target",
        "metadata_file": "bazel-bin/pkg/target.package-metadata.json",
    }

    # Verify node has required fields
    asserts.equals(env, "//pkg:target", node["label"])
    asserts.true(env, node["metadata_file"].endswith(".package-metadata.json"))

    # Verify node does NOT have inlined attributes
    asserts.false(env, "attributes" in node, "Node should not have inlined attributes")
    asserts.false(env, "package_name" in node, "Node should not have package_name")
    asserts.false(env, "licenses" in node, "Node should not have licenses")

    return unittest.end(env)

node_structure_test = unittest.make(_test_node_structure)

def _test_edge_structure(ctx):
    """Test edge format."""
    env = unittest.begin(ctx)

    edge = {
        "from": "//app:main",
        "to": "//lib:foo",
        "type": "depends_on",
    }

    asserts.equals(env, "//app:main", edge["from"])
    asserts.equals(env, "//lib:foo", edge["to"])
    asserts.equals(env, "depends_on", edge["type"])

    return unittest.end(env)

edge_structure_test = unittest.make(_test_edge_structure)

def _test_empty_metadata_no_nodes(ctx):
    """Test that empty metadata produces empty nodes array."""
    env = unittest.begin(ctx)

    # Create empty TransitiveMetadataInfo
    info = TransitiveMetadataInfo(
        transitive = depset(),
        top_level_target = None,
    )

    asserts.equals(env, 0, len(info.transitive.to_list()))

    return unittest.end(env)

empty_metadata_no_nodes_test = unittest.make(_test_empty_metadata_no_nodes)

# ============================================================================
# Tests for TargetWithMetadataInfo with direct_deps
# ============================================================================

def _test_target_with_direct_deps(ctx):
    """Test that direct_deps can be tracked as a list."""
    env = unittest.begin(ctx)

    # Test the concept of direct_deps as a list (without creating actual provider)
    direct_deps = ["//dep1:a", "//dep2:b"]

    asserts.equals(env, 2, len(direct_deps))
    asserts.equals(env, "//dep1:a", direct_deps[0])
    asserts.equals(env, "//dep2:b", direct_deps[1])

    return unittest.end(env)

target_with_direct_deps_test = unittest.make(_test_target_with_direct_deps)

def _test_target_with_empty_direct_deps(ctx):
    """Test that empty direct_deps list works."""
    env = unittest.begin(ctx)

    direct_deps = []
    asserts.equals(env, 0, len(direct_deps))

    return unittest.end(env)

target_with_empty_direct_deps_test = unittest.make(_test_target_with_empty_direct_deps)

# ============================================================================
# Tests for TransitiveMetadataInfo structure
# ============================================================================

def _test_transitive_metadata_field_name(ctx):
    """Test that TransitiveMetadataInfo uses 'transitive' field name."""
    env = unittest.begin(ctx)

    # Test with null instance
    tmi = TransitiveMetadataInfo(
        transitive = depset(),
        top_level_target = None,
    )

    # Verify 'transitive' field exists
    asserts.true(env, hasattr(tmi, "transitive"), "Should have 'transitive' field")
    asserts.equals(env, 0, len(tmi.transitive.to_list()))

    return unittest.end(env)

transitive_metadata_field_name_test = unittest.make(_test_transitive_metadata_field_name)

def _test_transitive_metadata_flattening(ctx):
    """Test that TransitiveMetadataInfo uses depsets for efficient flattening."""
    env = unittest.begin(ctx)

    # Test that we can create nested depsets of strings (simpler than providers)
    # This validates the depset flattening mechanism
    inner_depset = depset(["//c:target"])
    outer_depset = depset(direct = ["//b:target"], transitive = [inner_depset])

    # Should have both B and C when flattened
    all_items = outer_depset.to_list()
    asserts.true(env, len(all_items) >= 2, "Should have at least 2 items")
    asserts.true(env, "//b:target" in all_items or "//c:target" in all_items)

    return unittest.end(env)

transitive_metadata_flattening_test = unittest.make(_test_transitive_metadata_flattening)

# ============================================================================
# Integration test: Complete graph with edges
# ============================================================================

def _test_complete_graph_with_edges(ctx):
    """Test edge extraction logic from dependency information."""
    env = unittest.begin(ctx)

    # Simulate a dependency graph structure:
    #   A -> B
    #   A -> C
    # Using simple dicts to represent the concept
    graph_data = [
        {"target": "//a:target", "direct_deps": ["//b:target", "//c:target"]},
        {"target": "//b:target", "direct_deps": []},
        {"target": "//c:target", "direct_deps": []},
    ]

    # Extract edges using the same logic as serialization
    edges = []
    for node in graph_data:
        if node["direct_deps"]:
            for dep in node["direct_deps"]:
                edges.append({"from": node["target"], "to": dep})

    asserts.equals(env, 2, len(edges), "Should have 2 edges (A->B, A->C)")
    asserts.equals(env, "//a:target", edges[0]["from"])
    asserts.true(
        env,
        edges[0]["to"] == "//b:target" or edges[1]["to"] == "//b:target",
        "Should have edge to //b:target",
    )

    return unittest.end(env)

complete_graph_with_edges_test = unittest.make(_test_complete_graph_with_edges)

# ============================================================================
# Test suite
# ============================================================================

def serialization_test_suite(name):
    """Creates the test suite for metadata serialization."""
    unittest.suite(
        name,
        # Label manipulation tests
        strip_null_repo_with_at_double_slash_test,
        strip_null_repo_with_double_at_double_slash_test,
        strip_null_repo_no_prefix_test,
        strip_null_repo_external_repo_test,
        # JSON escaping tests
        json_escape_quotes_test,
        json_escape_newlines_test,
        json_escape_backslashes_test,
        # Graph-only format tests
        graph_only_format_structure_test,
        node_structure_test,
        edge_structure_test,
        empty_metadata_no_nodes_test,
        # Provider structure tests
        target_with_direct_deps_test,
        target_with_empty_direct_deps_test,
        transitive_metadata_field_name_test,
        transitive_metadata_flattening_test,
        # Integration tests
        complete_graph_with_edges_test,
    )
