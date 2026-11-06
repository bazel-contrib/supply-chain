"""Tests that official-case SPDX aliases resolve like lowercase targets."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load("@package_metadata//licenses:defs.bzl", "LicenseKindInfo")

def _license_kind_alias_test_impl(ctx):
    env = analysistest.begin(ctx)
    official = ctx.attr.official[LicenseKindInfo].identifier
    lowercase = ctx.attr.lowercase[LicenseKindInfo].identifier
    asserts.equals(env, ctx.attr.expected_identifier, official)
    asserts.equals(env, official, lowercase)
    return analysistest.end(env)

_license_kind_alias_test = analysistest.make(
    _license_kind_alias_test_impl,
    attrs = {
        "expected_identifier": attr.string(),
        "official": attr.label(providers = [LicenseKindInfo]),
        "lowercase": attr.label(providers = [LicenseKindInfo]),
    },
)

def spdx_alias_test(name, official, lowercase, identifier):
    """Checks that an official-case label and lowercase target provide one identifier."""
    _license_kind_alias_test(
        name = name,
        target_under_test = official,
        expected_identifier = identifier,
        official = official,
        lowercase = lowercase,
    )
