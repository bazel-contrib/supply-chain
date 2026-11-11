"""Tests of the `deprecation` field of `LicenseKindInfo`."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load("@package_metadata//licenses:defs.bzl", "LicenseKindInfo")

def _license_kind_test_impl(ctx):
    env = analysistest.begin(ctx)
    info = analysistest.target_under_test(env)[LicenseKindInfo]
    expected_deprecation = ctx.attr.expected_deprecation if ctx.attr.expected_deprecation else None
    asserts.equals(env, expected_deprecation, info.deprecation)
    return analysistest.end(env)

_license_kind_test = analysistest.make(
    _license_kind_test_impl,
    attrs = {
        "expected_deprecation": attr.string(),
    },
)

def license_kind_test(name, kind, deprecation = ""):
    """Tests the `deprecation` field of the `LicenseKindInfo` of a kind.

    Args:
      name: name of the test.
      kind: the `license_kind` target.
      deprecation: expected deprecation message, or the empty string if none.
    """
    _license_kind_test(
        name = name,
        target_under_test = kind,
        expected_deprecation = deprecation,
    )
