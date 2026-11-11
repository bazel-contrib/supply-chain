"""Tests of the `LicenseExceptionInfo` provider."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load("@package_metadata//licenses:defs.bzl", "LicenseExceptionInfo")

def _license_exception_test_impl(ctx):
    env = analysistest.begin(ctx)
    info = analysistest.target_under_test(env)[LicenseExceptionInfo]
    asserts.equals(env, ctx.attr.identifier, info.identifier)
    asserts.equals(env, ctx.attr.expected_name, info.name)
    expected_deprecation = ctx.attr.expected_deprecation if ctx.attr.expected_deprecation else None
    asserts.equals(env, expected_deprecation, info.deprecation)
    return analysistest.end(env)

_license_exception_test = analysistest.make(
    _license_exception_test_impl,
    attrs = {
        "expected_deprecation": attr.string(),
        "identifier": attr.string(),
        "expected_name": attr.string(),
    },
)

def license_exception_test(name, exception, identifier, exception_name, deprecation = ""):
    """Tests the provider fields of a `license_exception` target.

    Args:
      name: name of the test.
      exception: the `license_exception` target.
      identifier: expected identifier.
      exception_name: expected human-readable name.
      deprecation: expected deprecation message, or the empty string if none.
    """
    _license_exception_test(
        name = name,
        target_under_test = exception,
        identifier = identifier,
        expected_name = exception_name,
        expected_deprecation = deprecation,
    )
