"""Integration tests for serialization functions."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load(":gather_metadata.bzl", "gather_metadata_info")
load(":providers.bzl", "TransitiveMetadataInfo")
load(":serialization.bzl", "metadata_info_to_json")

def _json_output_structure_test_impl(ctx):
    """Verify metadata_info_to_json produces correct JSON structure."""
    env = analysistest.begin(ctx)

    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    json_output = metadata_info_to_json(info)

    asserts.equals(env, 1, len(json_output), "Should return list with one JSON string")

    parsed = json.decode(json_output[0])

    asserts.true(env, "schema_version" in parsed)
    asserts.equals(env, "1.0", parsed["schema_version"])
    asserts.true(env, "root_target" in parsed)
    asserts.true(env, "nodes" in parsed)
    asserts.true(env, "edges" in parsed)

    for node in parsed["nodes"]:
        asserts.true(env, "label" in node)
        asserts.true(env, "metadata_file" in node)

    for edge in parsed["edges"]:
        asserts.true(env, "from" in edge)
        asserts.true(env, "to" in edge)
        asserts.true(env, "type" in edge)
        asserts.equals(env, "depends_on", edge["type"])

    return analysistest.end(env)

json_output_structure_test = analysistest.make(
    _json_output_structure_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def _label_canonicalization_test_impl(ctx):
    """Verify label strings are canonical (no @ prefix for main workspace)."""
    env = analysistest.begin(ctx)

    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    json_output = metadata_info_to_json(info)
    parsed = json.decode(json_output[0])

    for node in parsed["nodes"]:
        label = node["label"]
        if not label.startswith("@"):
            asserts.true(
                env,
                label.startswith("//"),
                "Main workspace labels should start with // not @: {}".format(label),
            )

    return analysistest.end(env)

label_canonicalization_test = analysistest.make(
    _label_canonicalization_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def _edges_extracted_test_impl(ctx):
    """Verify edges are extracted from direct_deps."""
    env = analysistest.begin(ctx)

    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    json_output = metadata_info_to_json(info)
    parsed = json.decode(json_output[0])

    edges = parsed["edges"]

    if len(edges) > 0:
        for edge in edges:
            from_label = edge["from"]
            to_label = edge["to"]

            from_found = False
            to_found = False
            for node in parsed["nodes"]:
                if node["label"] == from_label:
                    from_found = True
                if node["label"] == to_label:
                    to_found = True

            asserts.true(env, from_found, "Edge 'from' should reference a node: {}".format(from_label))
            asserts.true(env, to_found, "Edge 'to' should reference a node: {}".format(to_label))

    return analysistest.end(env)

edges_extracted_test = analysistest.make(
    _edges_extracted_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def serialization_integration_test_suite(name):
    """Creates the serialization integration test suite."""

    json_output_structure_test(
        name = "json_output_structure_test",
        target_under_test = ":test_target_with_deps",
    )

    label_canonicalization_test(
        name = "label_canonicalization_test",
        target_under_test = ":test_target_with_deps",
    )

    edges_extracted_test(
        name = "edges_extracted_test",
        target_under_test = ":test_target_with_deps",
    )

    native.test_suite(
        name = name,
        tests = [
            ":json_output_structure_test",
            ":label_canonicalization_test",
            ":edges_extracted_test",
        ],
    )
