"""Small rules that declare toolchain types for SBOM integration tests."""

def _empty_rule_impl(ctx):
    return [DefaultInfo()]

def _toolchain_impl(ctx):
    return [platform_common.ToolchainInfo()]

toolchain_impl = rule(
    implementation = _toolchain_impl,
    attrs = {
        "applicable_licenses": attr.label_list(),
        "deps": attr.label_list(cfg = "exec"),
    },
)

java_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "@bazel_tools//tools/jdk:runtime_toolchain_type",
        "@bazel_tools//tools/jdk:toolchain_type",
    ],
)

py_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = ["@bazel_tools//tools/python:toolchain_type"],
)

js_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "//:node_runtime_type",
        "//:node_transpile_type",
    ],
)

cc_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = ["//:llvm_toolchain_type"],
)

openapi_generate_sources = rule(
    implementation = _empty_rule_impl,
    toolchains = ["@bazel_tools//tools/jdk:toolchain_type"],
)

py_image_layer = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "@bazel_tools//tools/python:toolchain_type",
        "//:tar_toolchain_type",
    ],
)

image_layer = rule(
    implementation = _empty_rule_impl,
    attrs = {
        "srcs": attr.label_list(),
        "provider_marker": attr.string(),
        "applicable_licenses": attr.label_list(),
        "_supply_chain_module": attr.string(default = Label("//:x").repo_name),
    },
)

provider_override_rule = rule(
    implementation = _empty_rule_impl,
    attrs = {
        "deps": attr.label_list(),
        "provider_marker": attr.string(),
        "applicable_licenses": attr.label_list(),
        "_supply_chain_module": attr.string(default = Label("//:x").repo_name),
    },
)

dual_use_image = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "//:dual_runtime_type",
        "@bazel_tools//tools/jdk:toolchain_type",
    ],
)

mixed_path_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "//:mixed_runtime_type",
        "//:mixed_build_type",
    ],
)

mixed_involvement_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "//:mixed_build_type",
        "//:mixed_embedded_type",
    ],
)

root_override_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = ["//:root_override_type"],
)
