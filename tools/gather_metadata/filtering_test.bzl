"""Tests for attribute filtering logic."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load(":gather_metadata.bzl", "gather_metadata_info")
load(":providers.bzl", "TransitiveMetadataInfo")

def _genrule_tools_filtered_test_impl(ctx):
    """Verify genrule 'tools' attribute is filtered (not traversed as dependency)."""
    env = analysistest.begin(ctx)

    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    all_targets = info.transitive.to_list()

    tool_label = str(ctx.label).replace("genrule_tools_filtered_test", "test_dep_b")

    edges_to_tool = []
    for target_info in all_targets:
        if hasattr(target_info, "direct_deps") and target_info.direct_deps:
            for dep in target_info.direct_deps:
                if str(dep) == tool_label:
                    edges_to_tool.append(str(target_info.target))

    asserts.equals(
        env,
        0,
        len(edges_to_tool),
        "Genrule 'tools' should be filtered, found edges: {}".format(edges_to_tool),
    )

    return analysistest.end(env)

genrule_tools_filtered_test = analysistest.make(
    _genrule_tools_filtered_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def _underscore_attrs_filtered_test_impl(ctx):
    """Verify attributes starting with _ are generally filtered."""
    env = analysistest.begin(ctx)

    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    all_targets = info.transitive.to_list()

    for target_info in all_targets:
        if hasattr(target_info, "direct_deps") and target_info.direct_deps:
            for dep in target_info.direct_deps:
                dep_str = str(dep)
                asserts.false(
                    env,
                    "_bash_binary" in dep_str or "_cc_toolchain" in dep_str,
                    "Should not traverse underscore attrs like _bash_binary",
                )

    return analysistest.end(env)

underscore_attrs_filtered_test = analysistest.make(
    _underscore_attrs_filtered_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def filtering_test_suite(name):
    """Creates the filtering test suite."""

    genrule_tools_filtered_test(
        name = "genrule_tools_filtered_test",
        target_under_test = ":test_genrule_with_tools",
    )

    underscore_attrs_filtered_test(
        name = "underscore_attrs_filtered_test",
        target_under_test = ":test_target_with_deps",
    )

    native.test_suite(
        name = name,
        tests = [
            ":genrule_tools_filtered_test",
            ":underscore_attrs_filtered_test",
        ],
    )
