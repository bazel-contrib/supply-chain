"""Tests that --//sbom:strict_validations controls the validators of the exporters."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")

def _validator_args(env):
    for action in analysistest.target_actions(env):
        if "--graph" in action.argv:
            return [arg for arg in action.argv if arg == "--validator"]
    fail("no exporter action found")

def _validates_test_impl(ctx):
    env = analysistest.begin(ctx)
    asserts.equals(env, ["--validator"], _validator_args(env))
    return analysistest.end(env)

def _does_not_validate_test_impl(ctx):
    env = analysistest.begin(ctx)
    asserts.equals(env, [], _validator_args(env))
    return analysistest.end(env)

_strict_flag = str(Label("@supply_chain_tools//sbom:strict_validations"))

_validates_strict_test = analysistest.make(
    _validates_test_impl,
    config_settings = {_strict_flag: True},
)

_does_not_validate_default_test = analysistest.make(
    _does_not_validate_test_impl,
    config_settings = {_strict_flag: False},
)

def strict_validations_test_suite(name, cyclonedx, spdx):
    """Tests of --//sbom:strict_validations.

    Args:
      name: name of the test suite.
      cyclonedx: a `cyclonedx` target.
      spdx: an `spdx` target.
    """
    _validates_strict_test(name = name + "_cyclonedx_strict", target_under_test = cyclonedx)
    _does_not_validate_default_test(name = name + "_cyclonedx_default", target_under_test = cyclonedx)

    _validates_strict_test(name = name + "_spdx_strict", target_under_test = spdx)
    _does_not_validate_default_test(name = name + "_spdx_default", target_under_test = spdx)

    native.test_suite(
        name = name,
        tests = [
            name + "_cyclonedx_strict",
            name + "_cyclonedx_default",
            name + "_spdx_strict",
            name + "_spdx_default",
        ],
    )
