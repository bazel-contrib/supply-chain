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
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("@bazel_tools//tools/jdk:toolchain_type"): "build",
            Label("@bazel_tools//tools/jdk:runtime_toolchain_type"): "runtime",
        }),
    },
)

py_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = ["@bazel_tools//tools/python:toolchain_type"],
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("@bazel_tools//tools/python:toolchain_type"): "runtime",
        }),
    },
)

js_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "//:node_runtime_type",
        "//:node_transpile_type",
    ],
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//:node_runtime_type"): "runtime",
            Label("//:node_transpile_type"): "build",
        }),
    },
)

cc_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = ["//:llvm_toolchain_type"],
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//:llvm_toolchain_type"): "embedded",
        }),
    },
)

openapi_generate_sources = rule(
    implementation = _empty_rule_impl,
    toolchains = ["@bazel_tools//tools/jdk:toolchain_type"],
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("@bazel_tools//tools/jdk:toolchain_type"): "build",
        }),
    },
)

py_image_layer = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "@bazel_tools//tools/python:toolchain_type",
        "//:tar_toolchain_type",
    ],
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("@bazel_tools//tools/python:toolchain_type"): "embedded",
            Label("//:tar_toolchain_type"): "build",
        }),
    },
)

image_layer = rule(
    implementation = _empty_rule_impl,
    attrs = {
        "srcs": attr.label_list(),
        "provider_marker": attr.string(),
        "applicable_licenses": attr.label_list(),
        "_supply_chain_relationships": attr.string_dict(default = {
            "srcs": "dynamic_link",
        }),
    },
)

provider_override_rule = rule(
    implementation = _empty_rule_impl,
    attrs = {
        "deps": attr.label_list(),
        "provider_marker": attr.string(),
        "applicable_licenses": attr.label_list(),
        "_supply_chain_relationships": attr.string_dict(default = {
            "deps": "static_link",
        }),
    },
)

dual_use_image = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "//:dual_runtime_type",
        "@bazel_tools//tools/jdk:toolchain_type",
    ],
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//:dual_runtime_type"): "runtime",
            Label("@bazel_tools//tools/jdk:toolchain_type"): "build",
        }),
    },
)

mixed_path_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "//:mixed_runtime_type",
        "//:mixed_build_type",
    ],
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//:mixed_runtime_type"): "runtime",
            Label("//:mixed_build_type"): "build",
        }),
    },
)

mixed_involvement_binary = rule(
    implementation = _empty_rule_impl,
    toolchains = [
        "//:mixed_build_type",
        "//:mixed_embedded_type",
    ],
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//:mixed_build_type"): "build",
            Label("//:mixed_embedded_type"): "embedded",
        }),
    },
)
