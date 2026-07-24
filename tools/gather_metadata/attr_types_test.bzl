"""Tests that dependency edges are gathered through every attribute type.

The metadata gathering aspect must follow dependency edges regardless of the
attribute type that declares them: plain ``label`` attributes, ``label_list``
attributes, ``label_keyed_string_dict`` attributes (targets in the dict keys)
and ``string_keyed_label_dict`` attributes (targets in the dict values). These
tests define a small rule for each attribute shape and assert that a
metadata-carrying dependency reached *only* through that attribute is captured
in the resulting TransitiveMetadataInfo (both as a transitive node and as a
direct edge).
"""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load(":gather_metadata.bzl", "gather_metadata_info")
load(":providers.bzl", "TransitiveMetadataInfo")

def _consume_impl(_ctx):
    return [DefaultInfo()]

# A rule per attribute shape that can carry dependency edges. Each simply
# depends on its input(s); the metadata is attached via the package's
# default_package_metadata, so any traversed dependency shows up in the graph.
_label_attr_rule = rule(
    implementation = _consume_impl,
    attrs = {"dep": attr.label()},
)

_label_list_attr_rule = rule(
    implementation = _consume_impl,
    attrs = {"deps": attr.label_list()},
)

_label_keyed_string_dict_attr_rule = rule(
    implementation = _consume_impl,
    attrs = {"deps": attr.label_keyed_string_dict()},
)

_string_keyed_label_dict_attr_rule = rule(
    implementation = _consume_impl,
    attrs = {"deps": attr.string_keyed_label_dict()},
)

def _run_check(ctx, dep_suffix):
    """Asserts the dependency `:dep_suffix` was gathered by the aspect."""
    env = analysistest.begin(ctx)
    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    all_targets = info.transitive.to_list()
    labels = [str(twmi.target) for twmi in all_targets]

    # 1. The dependency reached through the attribute must appear as a node in
    #    the transitive metadata set.
    reached = [label for label in labels if label.endswith(":" + dep_suffix)]
    asserts.true(
        env,
        len(reached) > 0,
        "Expected dependency ':{}' to be gathered through the attribute, got {}".format(
            dep_suffix,
            labels,
        ),
    )

    # 2. The edge from the target under test to that dependency must be recorded
    #    in direct_deps for graph/SBOM generation.
    direct_deps = None
    for twmi in all_targets:
        if twmi.target == target_under_test.label:
            direct_deps = [str(dep) for dep in twmi.direct_deps]

    asserts.true(
        env,
        direct_deps != None,
        "Target under test should have its own metadata entry",
    )
    if direct_deps != None:
        edges = [dep for dep in direct_deps if dep.endswith(":" + dep_suffix)]
        asserts.true(
            env,
            len(edges) > 0,
            "Expected a direct edge to ':{}', got {}".format(dep_suffix, direct_deps),
        )

    return analysistest.end(env)

def _via_label_test_impl(ctx):
    return _run_check(ctx, "label_leaf")

def _via_label_list_test_impl(ctx):
    return _run_check(ctx, "label_list_leaf")

def _via_dict_test_impl(ctx):
    return _run_check(ctx, "dict_leaf")

def _via_string_keyed_dict_test_impl(ctx):
    return _run_check(ctx, "string_keyed_dict_leaf")

via_label_test = analysistest.make(
    _via_label_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

via_label_list_test = analysistest.make(
    _via_label_list_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

via_dict_test = analysistest.make(
    _via_dict_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

via_string_keyed_dict_test = analysistest.make(
    _via_string_keyed_dict_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def attr_types_test_suite(name):
    """Creates the per-attribute-type traversal test suite.

    Args:
      name: the name of the generated test_suite target.
    """

    # Each leaf is reachable exclusively through one attribute type, so a test
    # passing proves that specific attribute type is traversed.
    native.filegroup(name = "label_leaf", srcs = [], tags = ["manual"])
    native.filegroup(name = "label_list_leaf", srcs = [], tags = ["manual"])
    native.filegroup(name = "dict_leaf", srcs = [], tags = ["manual"])
    native.filegroup(name = "string_keyed_dict_leaf", srcs = [], tags = ["manual"])

    _label_attr_rule(
        name = "via_label",
        dep = ":label_leaf",
        tags = ["manual"],
    )
    _label_list_attr_rule(
        name = "via_label_list",
        deps = [":label_list_leaf"],
        tags = ["manual"],
    )
    _label_keyed_string_dict_attr_rule(
        name = "via_dict",
        deps = {":dict_leaf": "unused-value"},
        tags = ["manual"],
    )
    _string_keyed_label_dict_attr_rule(
        name = "via_string_keyed_dict",
        deps = {"unused-key": ":string_keyed_dict_leaf"},
        tags = ["manual"],
    )

    via_label_test(
        name = "via_label_test",
        target_under_test = ":via_label",
    )
    via_label_list_test(
        name = "via_label_list_test",
        target_under_test = ":via_label_list",
    )
    via_dict_test(
        name = "via_dict_test",
        target_under_test = ":via_dict",
    )
    via_string_keyed_dict_test(
        name = "via_string_keyed_dict_test",
        target_under_test = ":via_string_keyed_dict",
    )

    native.test_suite(
        name = name,
        tests = [
            ":via_label_test",
            ":via_label_list_test",
            ":via_dict_test",
            ":via_string_keyed_dict_test",
        ],
    )
