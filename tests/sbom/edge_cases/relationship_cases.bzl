"""Fixture rules with image/package/link attribute shapes for relationship tests."""

load("@supply_chain_tools//sbom:cyclonedx.bzl", "cyclonedx")
load("@supply_chain_tools//sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//sbom:spdx.bzl", "spdx")
load("@package_metadata_extensions//providers:package_linking_info.bzl", "PackageLinkingInfo")
load("//sbom:golden_test.bzl", "sbom_golden_test")

def _fixture_impl(ctx):
    return [DefaultInfo()]

image_layer = rule(
    implementation = _fixture_impl,
    attrs = {
        "srcs": attr.label_list(),
        "root_marker": attr.string(),
        "applicable_licenses": attr.label_list(),
        "_supply_chain_relationships": attr.string_dict(default = {
            "srcs": "contains",
        }),
    },
)

root_image_layer = rule(
    implementation = _fixture_impl,
    attrs = {
        "srcs": attr.label_list(),
        "root_marker": attr.string(),
        "applicable_licenses": attr.label_list(),
        "_supply_chain_relationships": attr.string_dict(default = {
            "srcs": "static_link",
        }),
    },
)

def _linking_root_impl(ctx):
    static_names = {name: True for name in ctx.attr.static_names}
    dynamic_names = {name: True for name in ctx.attr.dynamic_names}
    return [
        DefaultInfo(),
        PackageLinkingInfo(
            static_linked = [dep.label for dep in ctx.attr.deps if dep.label.name in static_names],
            dynamic_linked = [dep.label for dep in ctx.attr.deps if dep.label.name in dynamic_names],
        ),
    ]

linking_root = rule(
    implementation = _linking_root_impl,
    attrs = {
        "deps": attr.label_list(),
        "static_names": attr.string_list(),
        "dynamic_names": attr.string_list(),
        "applicable_licenses": attr.label_list(),
        "_supply_chain_relationships": attr.string_dict(default = {
            "deps": "contains",
        }),
    },
)

pkg_tar_impl = rule(
    implementation = _fixture_impl,
    attrs = {
        "srcs": attr.label_list(),
        "deps": attr.label_list(),
        "applicable_licenses": attr.label_list(),
        "_supply_chain_relationships": attr.string_dict(default = {
            "srcs": "contains",
            "deps": "contains",
        }),
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
        "_supply_chain_relationships": attr.string_dict(default = {
            "layers": "contains",
            "static_deps": "static_link",
            "dynamic_deps": "dynamic_link",
            "test_deps": "test_dependency",
            "dev_deps": "dev_dependency",
        }),
    },
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
