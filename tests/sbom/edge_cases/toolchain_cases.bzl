"""Golden tests for toolchain relationships in generated SBOMs."""

load("@bazel_features//:features.bzl", "bazel_features")
load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
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

_ToolchainAliasProbeInfo = provider(fields = ["key_label"])

def _toolchain_alias_probe_aspect_impl(_target, ctx):
    key_target = ctx.rule.attr._supply_chain_toolchain_involvement.keys()[0]
    return [_ToolchainAliasProbeInfo(
        key_label = key_target.label,
    )]

_toolchain_alias_probe_aspect = aspect(
    implementation = _toolchain_alias_probe_aspect_impl,
)

def _toolchain_alias_probe_rule_impl(_ctx):
    return [DefaultInfo()]

_toolchain_alias_probe_rule = rule(
    implementation = _toolchain_alias_probe_rule_impl,
    toolchains = ["//sbom/edge_cases:cycle_backedge_type_alias"],
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//sbom/edge_cases:cycle_backedge_type_alias"): "runtime",
        }),
    },
)

def _toolchain_alias_probe_test_impl(ctx):
    env = analysistest.begin(ctx)
    probe = analysistest.target_under_test(env)[_ToolchainAliasProbeInfo]
    asserts.equals(
        env,
        Label("//sbom/edge_cases:cycle_backedge_type"),
        probe.key_label,
        "The label-keyed attribute should expose the actual toolchain type target",
    )
    return analysistest.end(env)

_toolchain_alias_probe_test = analysistest.make(
    _toolchain_alias_probe_test_impl,
    extra_target_under_test_aspects = [_toolchain_alias_probe_aspect],
)

def toolchain_alias_probe_test(name):
    _toolchain_alias_probe_rule(name = name + "_target")
    _toolchain_alias_probe_test(
        name = name,
        target_under_test = ":" + name + "_target",
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
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//sbom/edge_cases:cycle_backedge_type"): "runtime",
        }),
    },
)

def _invalid_toolchain_attr_impl(ctx):
    return [DefaultInfo()]

_invalid_value_toolchain_rule = rule(
    implementation = _invalid_toolchain_attr_impl,
    toolchains = ["//sbom/edge_cases:cycle_backedge_type"],
    attrs = {
        "applicable_licenses": attr.label_list(),
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//sbom/edge_cases:cycle_backedge_type"): "invalid",
        }),
    },
)

_invalid_key_toolchain_rule = rule(
    implementation = _invalid_toolchain_attr_impl,
    toolchains = ["//sbom/edge_cases:cycle_backedge_type"],
    attrs = {
        "applicable_licenses": attr.label_list(),
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//sbom/edge_cases:toolchain_app_metadata"): "runtime",
        }),
    },
)

_invalid_type_toolchain_rule = rule(
    implementation = _invalid_toolchain_attr_impl,
    toolchains = ["//sbom/edge_cases:cycle_backedge_type"],
    attrs = {
        "applicable_licenses": attr.label_list(),
        "_supply_chain_toolchain_involvement": attr.string(default = "invalid"),
    },
)

_invalid_relationship_value_rule = rule(
    implementation = _invalid_toolchain_attr_impl,
    attrs = {
        "applicable_licenses": attr.label_list(),
        "deps": attr.label_list(),
        "_supply_chain_relationships": attr.string_dict(default = {
            "deps": "invalid",
        }),
    },
)

_invalid_relationship_type_rule = rule(
    implementation = _invalid_toolchain_attr_impl,
    attrs = {
        "applicable_licenses": attr.label_list(),
        "_supply_chain_relationships": attr.string(default = "invalid"),
    },
)

def _invalid_toolchain_attr_test_impl(ctx):
    env = analysistest.begin(ctx)
    asserts.expect_failure(env, ctx.attr.expected_message)
    return analysistest.end(env)

_invalid_toolchain_attr_test = analysistest.make(
    _invalid_toolchain_attr_test_impl,
    attrs = {"expected_message": attr.string(mandatory = True)},
    expect_failure = True,
)

def invalid_toolchain_attr_tests(name, applicable_licenses):
    """Test that the aspect rejects invalid rule-declared attribute data."""
    tests = []
    for suffix, target_rule, message in [
        ("invalid_toolchain_key", _invalid_key_toolchain_rule, "Invalid _supply_chain_toolchain_involvement key"),
        ("invalid_toolchain_value", _invalid_value_toolchain_rule, "Invalid _supply_chain_toolchain_involvement value"),
        ("invalid_toolchain_type", _invalid_type_toolchain_rule, "invalid _supply_chain_toolchain_involvement attribute type"),
        ("invalid_relationship_value", _invalid_relationship_value_rule, "Invalid _supply_chain_relationships value"),
        ("invalid_relationship_type", _invalid_relationship_type_rule, "invalid _supply_chain_relationships attribute type"),
    ]:
        target_name = name + "_" + suffix + "_target"
        sbom_name = name + "_" + suffix + "_sbom"
        test_name = name + "_" + suffix + "_test"
        target_rule(
            name = target_name,
            applicable_licenses = [applicable_licenses],
            tags = ["manual"],
        )
        sbom(
            name = sbom_name,
            tags = ["manual"],
            target = ":" + target_name,
        )
        _invalid_toolchain_attr_test(
            name = test_name,
            target_under_test = ":" + sbom_name,
            expected_message = message,
        )
        tests.append(":" + test_name)
    native.test_suite(name = name, tests = tests)

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
