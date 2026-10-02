"""Tests that a dependency reached more than once from one target produces a single edge.

A single dependency can be reachable from its parent through more than one
attribute (e.g. both "deps" and "embed" naming the same library, which is
common for generated third-party Go targets), or appear as the value of more
than one entry in a single dict-typed attribute (e.g. "importmap"-style
attributes keyed by something other than the label itself; unlike
label_list/label_keyed_string_dict, Bazel does not require dict values to be
unique). Before the fix for
https://github.com/bazel-contrib/supply-chain/issues/214, each such path
independently contributed its own `depends_on` edge and transitive metadata
slice, so SBOM output for real packages showed every child dependency
duplicated (and, once two such duplicate-labeled nodes were later collapsed
onto one SBOM component, a dependency edge between them turned into a
self-reference).

These tests pin that a single dependency produces exactly one direct_deps
entry and exactly one serialized graph edge, no matter how many paths reach
it. Note that Bazel itself rejects a literal duplicate label within a single
label_list/label_keyed_string_dict attribute ("Label '...' is duplicated in
the '...' attribute"), so that specific shape can't occur and isn't tested
here; the dict-value case below is the realistic analogue of reaching one
dependency twice through a single attribute.
"""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load(":gather_metadata.bzl", "gather_metadata_info")
load(":providers.bzl", "TransitiveMetadataInfo")
load(":serialization.bzl", "metadata_info_to_json")

def _consume_impl(_ctx):
    return [DefaultInfo()]

# A dependency reachable through two distinct label_list attributes on the
# same rule (e.g. "deps" and "embed" on a real go_library).
_dual_attr_rule = rule(
    implementation = _consume_impl,
    attrs = {
        "deps": attr.label_list(),
        "extra_deps": attr.label_list(),
    },
)

# A dependency reachable twice through the *values* of a single
# string_keyed_label_dict attribute. Dict keys must be unique, but Bazel does
# not require values to be unique, so the same target can legitimately appear
# under two different keys.
_string_keyed_dict_rule = rule(
    implementation = _consume_impl,
    attrs = {
        "deps": attr.string_keyed_label_dict(),
    },
)

def _single_edge_per_dependency_check(ctx, dep_suffix):
    """Asserts `dep_suffix` produces exactly one direct edge from the target under test."""
    env = analysistest.begin(ctx)
    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    all_targets = info.transitive.to_list()

    # 1. direct_deps on the target's own TargetWithMetadataInfo must list the
    #    dependency exactly once.
    direct_deps = None
    for twmi in all_targets:
        if twmi.target == target_under_test.label:
            direct_deps = [str(dep) for dep in twmi.direct_deps]

    asserts.true(env, direct_deps != None, "Target under test should have its own metadata entry")

    if direct_deps != None:
        matches = [dep for dep in direct_deps if dep.endswith(":" + dep_suffix)]
        asserts.equals(
            env,
            1,
            len(matches),
            "Expected exactly one direct_deps entry for ':{}', got {}".format(dep_suffix, direct_deps),
        )

    # 2. The same must hold once serialized into the graph JSON that the SBOM
    #    generators consume directly. Match "from" by name suffix rather than
    #    comparing against str(target_under_test.label) directly: edges are
    #    built through _label_to_string(), which normalizes main-workspace
    #    labels to bare "//pkg:name", while str(Label) varies across Bazel
    #    versions (e.g. the "@@" bzlmod prefix), so a direct comparison would
    #    never match.
    from_suffix = ":" + target_under_test.label.name
    json_output = metadata_info_to_json(info)
    parsed = json.decode(json_output[0])
    edges_to_dep = [
        edge
        for edge in parsed["edges"]
        if edge["to"].endswith(":" + dep_suffix) and edge["from"].endswith(from_suffix)
    ]
    asserts.equals(
        env,
        1,
        len(edges_to_dep),
        "Expected exactly one serialized edge to ':{}', got {}".format(dep_suffix, edges_to_dep),
    )

    return analysistest.end(env)

def _via_two_attrs_test_impl(ctx):
    return _single_edge_per_dependency_check(ctx, "dedup_shared_leaf")

def _via_repeated_dict_value_test_impl(ctx):
    return _single_edge_per_dependency_check(ctx, "dedup_shared_leaf")

via_two_attrs_test = analysistest.make(
    _via_two_attrs_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

via_repeated_dict_value_test = analysistest.make(
    _via_repeated_dict_value_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def dedup_test_suite(name):
    """Creates the dependency-deduplication test suite.

    Args:
      name: the name of the generated test_suite target.
    """

    native.filegroup(name = "dedup_shared_leaf", srcs = [], tags = ["manual"])

    # The same dependency, reached via two different attributes.
    _dual_attr_rule(
        name = "dedup_via_two_attrs",
        deps = [":dedup_shared_leaf"],
        extra_deps = [":dedup_shared_leaf"],
        tags = ["manual"],
    )

    # The same dependency, appearing as the value of two different keys
    # within a single string_keyed_label_dict attribute.
    _string_keyed_dict_rule(
        name = "dedup_via_repeated_dict_value",
        deps = {
            "first": ":dedup_shared_leaf",
            "second": ":dedup_shared_leaf",
        },
        tags = ["manual"],
    )

    via_two_attrs_test(
        name = "dedup_via_two_attrs_test",
        target_under_test = ":dedup_via_two_attrs",
    )

    via_repeated_dict_value_test(
        name = "dedup_via_repeated_dict_value_test",
        target_under_test = ":dedup_via_repeated_dict_value",
    )

    native.test_suite(
        name = name,
        tests = [
            ":dedup_via_two_attrs_test",
            ":dedup_via_repeated_dict_value_test",
        ],
    )
