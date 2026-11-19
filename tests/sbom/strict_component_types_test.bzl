"""Tests that --//sbom:strict_component_types fails analysis for untyped components."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")

_strict_flag = str(Label("@supply_chain_tools//sbom:strict_component_types"))

def _fails_test_impl(ctx):
    env = analysistest.begin(ctx)
    asserts.expect_failure(env, "without a component_type attribute")
    asserts.expect_failure(env, "//sbom/edge_cases:image_library")
    return analysistest.end(env)

def _passes_test_impl(ctx):
    env = analysistest.begin(ctx)
    return analysistest.end(env)

_fails_strict_test = analysistest.make(
    _fails_test_impl,
    expect_failure = True,
    config_settings = {_strict_flag: True},
)

_passes_strict_test = analysistest.make(
    _passes_test_impl,
    config_settings = {_strict_flag: True},
)

_passes_default_test = analysistest.make(
    _passes_test_impl,
    config_settings = {_strict_flag: False},
)

def strict_component_types_test_suite(name, untyped_sbom, typed_sbom):
    """Tests of --//sbom:strict_component_types.

    Args:
      name: name of the test suite.
      untyped_sbom: an `sbom` target whose graph has a component without a component_type.
      typed_sbom: an `sbom` target whose components all declare a component_type.
    """
    _fails_strict_test(name = name + "_untyped_strict", target_under_test = untyped_sbom)
    _passes_default_test(name = name + "_untyped_default", target_under_test = untyped_sbom)
    _passes_strict_test(name = name + "_typed_strict", target_under_test = typed_sbom)

    native.test_suite(
        name = name,
        tests = [
            name + "_untyped_strict",
            name + "_untyped_default",
            name + "_typed_strict",
        ],
    )
