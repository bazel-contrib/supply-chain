"""Tests of the `deprecated` and `exception` fields of `LicenseKindInfo`."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load("@package_metadata//licenses:defs.bzl", "LicenseKindInfo")

def _license_kind_test_impl(ctx):
    env = analysistest.begin(ctx)
    info = analysistest.target_under_test(env)[LicenseKindInfo]
    asserts.equals(env, ctx.attr.deprecated, info.deprecated)
    asserts.equals(env, ctx.attr.exception, info.exception)
    return analysistest.end(env)

_license_kind_test = analysistest.make(
    _license_kind_test_impl,
    attrs = {
        "deprecated": attr.bool(),
        "exception": attr.bool(),
    },
)

def license_kind_test(name, kind, deprecated = False, exception = False):
    """Tests the `deprecated` and `exception` fields of the `LicenseKindInfo` of a kind.

    Args:
      name: name of the test.
      kind: the `license_kind` target.
      deprecated: expected value of `deprecated`.
      exception: expected value of `exception`.
    """
    _license_kind_test(
        name = name,
        target_under_test = kind,
        deprecated = deprecated,
        exception = exception,
    )
