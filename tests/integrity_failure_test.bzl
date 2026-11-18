"""Analysis tests for `integrity` targets with invalid values."""

load("@package_metadata//attributes:integrity.bzl", "integrity")
load("@rules_testing//lib:analysis_test.bzl", "analysis_test")
load("@rules_testing//lib:truth.bzl", "matching")

def integrity_failure_test(*, name, values, message):
    """Asserts that analyzing an `integrity` with `values` fails with `message`."""
    integrity(
        name = name + "_integrity",
        values = values,
        tags = ["manual"],
    )
    analysis_test(
        name = name,
        impl = lambda env, target: _failure_impl(env, target, message),
        target = name + "_integrity",
        expect_failure = True,
    )

def _failure_impl(env, target, message):
    env.expect.that_target(target).failures().contains_predicate(
        matching.str_matches("*" + message + "*"),
    )
