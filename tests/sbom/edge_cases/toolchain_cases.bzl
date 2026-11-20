"""Golden tests for toolchain relationships in generated SBOMs."""

load("@bazel_features//:features.bzl", "bazel_features")
load("@supply_chain_tools//sbom:cyclonedx.bzl", "cyclonedx")
load("@supply_chain_tools//sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//sbom:spdx.bzl", "spdx")
load("//sbom:golden_test.bzl", "sbom_golden_test")

_BUILD_TOOLS_TRANSITION = transition(
    implementation = lambda settings, attr: {
        "@supply_chain_tools//sbom:include_build_tools": True,
    },
    inputs = [],
    outputs = ["@supply_chain_tools//sbom:include_build_tools"],
)

def _configured_target_impl(ctx):
    return [DefaultInfo()]

def _cycle_backedge_binary_impl(ctx):
    return [DefaultInfo()]

_configured_target = rule(
    implementation = _configured_target_impl,
    attrs = {
        "dep": attr.label(mandatory = True, cfg = _BUILD_TOOLS_TRANSITION),
        "applicable_licenses": attr.label_list(),
    },
)

cycle_backedge_binary = rule(
    implementation = _cycle_backedge_binary_impl,
    toolchains = ["//sbom/edge_cases:cycle_backedge_type"],
)

def toolchain_golden_cases(name, cases, applicable_licenses):
    """Create three format goldens for each toolchain fixture.

    Toolchain-aspect support was added in Bazel 8.8. Older 8.x CI still runs
    the rest of the test suite, but cannot observe these edges.
    """
    if not bazel_features.rules.supports_toolchains_aspects_star:
        return

    tests = []
    for case in cases:
        case_name = case["name"]
        configured_name = case_name + "_configured"
        _configured_target(
            name = configured_name,
            dep = case["target"],
            applicable_licenses = applicable_licenses,
        )

        sbom_name = case_name + "_sbom"
        sbom(
            name = sbom_name,
            target = ":" + configured_name,
        )

        reports = [
            ("cyclonedx", "cyclonedx"),
            ("spdx23", "spdx"),
            ("spdx301", "spdx"),
        ]
        for suffix, kind in reports:
            report_name = case_name + "_" + suffix
            if kind == "cyclonedx":
                cyclonedx(
                    name = report_name,
                    sbom = ":" + sbom_name,
                )
            else:
                spdx(
                    name = report_name,
                    format = "json-ld" if suffix == "spdx301" else "json",
                    sbom = ":" + sbom_name,
                )

            test_name = report_name + "_golden_test"
            sbom_golden_test(
                name = test_name,
                golden = "testdata/" + report_name + ".json",
                report = ":" + report_name,
            )
            tests.append(":" + test_name)

    native.test_suite(
        name = name,
        tests = tests,
    )

def mixed_path_cyclonedx_golden(name, target, applicable_licenses):
    """Test that a shared PURL remains in scope when one path is runtime."""
    configured_name = name + "_configured"
    _configured_target(
        name = configured_name,
        dep = target,
        applicable_licenses = applicable_licenses,
    )

    sbom_name = name + "_sbom"
    sbom(
        name = sbom_name,
        target = ":" + configured_name,
    )

    report_name = name + "_cyclonedx"
    cyclonedx(
        name = report_name,
        sbom = ":" + sbom_name,
    )

    sbom_golden_test(
        name = report_name + "_golden_test",
        golden = "testdata/" + report_name + ".json",
        report = ":" + report_name,
    )
