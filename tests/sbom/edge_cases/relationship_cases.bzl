"""Fixture rules with image/package/link attribute shapes for relationship tests."""

load("@rules_testing//lib:analysis_test.bzl", "analysis_test")
load("@rules_testing//lib:truth.bzl", "matching")
load("@supply_chain_tools//sbom:cyclonedx.bzl", "cyclonedx")
load("@supply_chain_tools//sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//sbom:spdx.bzl", "spdx")
load("//sbom:golden_test.bzl", "sbom_golden_test")

def _fixture_impl(ctx):
    return [DefaultInfo()]

image_layer = rule(
    implementation = _fixture_impl,
    attrs = {
        "srcs": attr.label_list(),
        "root_marker": attr.string(),
        "applicable_licenses": attr.label_list(),
    },
)

pkg_tar_impl = rule(
    implementation = _fixture_impl,
    attrs = {
        "srcs": attr.label_list(),
        "deps": attr.label_list(),
        "applicable_licenses": attr.label_list(),
    },
)

ambiguous_fixture = rule(
    implementation = _fixture_impl,
    attrs = {
        "deps": attr.label_list(),
        "ambiguity_left": attr.string(),
        "ambiguity_right": attr.string(),
        "applicable_licenses": attr.label_list(),
    },
)

relationship_root = rule(
    implementation = _fixture_impl,
    attrs = {
        "layers": attr.label_list(),
        "deps": attr.label_list(),
        "static_deps": attr.label_list(),
        "dynamic_deps": attr.label_list(),
        "test_deps": attr.label_list(),
        "dev_deps": attr.label_list(),
        "legacy_deps": attr.label_list(),
        "applicable_licenses": attr.label_list(),
    },
)

def ambiguous_relationship_failure_test(name, target):
    """Asserts that equally specific conflicting declarations fail analysis."""
    sbom(
        name = name + "_sbom",
        target = target,
        tags = ["manual"],
    )
    analysis_test(
        name = name,
        impl = lambda env, target: env.expect.that_target(target).failures().contains_predicate(
            matching.str_matches("*rule kind 'ambiguous_fixture' attribute 'deps'*contains*declared by module 'supply_chain_tests'*static_link*declared by module 'supply_chain_tests'*"),
        ),
        target = ":" + name + "_sbom",
        expect_failure = True,
    )

def relationship_golden_tests(name, target):
    """Create the three SBOM goldens for the relationship fixture graph."""
    sbom(name = name + "_sbom", target = target)
    reports = [
        ("cyclonedx", "cyclonedx", "json"),
        ("spdx23", "spdx", "json"),
        ("spdx301", "spdx", "json-ld"),
    ]
    tests = []
    for suffix, kind, output_format in reports:
        report_name = name + "_" + suffix
        if kind == "cyclonedx":
            cyclonedx(name = report_name, sbom = ":" + name + "_sbom")
        else:
            spdx(name = report_name, format = output_format, sbom = ":" + name + "_sbom")
        test_name = report_name + "_golden_test"
        sbom_golden_test(
            name = test_name,
            golden = "testdata/" + report_name + ".json",
            report = ":" + report_name,
        )
        tests.append(":" + test_name)
    native.test_suite(name = name, tests = tests)
