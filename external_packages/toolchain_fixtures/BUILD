load(":fixtures.bzl", "image_layer", "provider_override_rule", "toolchain_impl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package(default_visibility = ["//visibility:public"])

toolchain_type(name = "node_runtime_type")
toolchain_type(name = "node_transpile_type")
toolchain_type(name = "tar_toolchain_type")
toolchain_type(name = "dual_runtime_type")
toolchain_type(name = "mixed_runtime_type")
toolchain_type(name = "mixed_build_type")
toolchain_type(name = "mixed_embedded_type")
toolchain_type(name = "llvm_toolchain_type")

package_metadata(
    name = "provider_image_layer_metadata",
    purl = "pkg:generic/provider-image-layer@1.0.0",
)

package_metadata(
    name = "provider_override_metadata",
    purl = "pkg:generic/provider-override@1.0.0",
)

package_metadata(
    name = "provider_payload_metadata",
    purl = "pkg:generic/server-binary@1.0.0",
)

filegroup(
    name = "provider_payload",
    applicable_licenses = [":provider_payload_metadata"],
)

image_layer(
    name = "provider_image_layer",
    srcs = [":provider_payload"],
    applicable_licenses = [":provider_image_layer_metadata"],
)

provider_override_rule(
    name = "provider_override",
    applicable_licenses = [":provider_override_metadata"],
    deps = [":provider_payload"],
)

[
    package_metadata(
        name = name + "_metadata",
        purl = "pkg:generic/toolchain-fixtures-" + name + "@1.0.0",
    )
    for name in [
        "java_jre",
        "javac",
        "python",
        "node_runtime",
        "node_transpiler",
        "llvm_runtime",
        "image_tar",
    ]
]

[
    package_metadata(
        name = name + "_metadata",
        purl = "pkg:generic/toolchain-fixtures-shared-tool@1.0.0",
    )
    for name in [
        "mixed_build",
        "mixed_runtime",
        "mixed_embedded",
    ]
]

[
    toolchain_impl(
        name = name,
        applicable_licenses = [":" + name + "_metadata"],
    )
    for name in [
        "mixed_build",
        "mixed_runtime",
        "mixed_embedded",
    ]
]

[
    toolchain_impl(
        name = name,
        applicable_licenses = [":" + name + "_metadata"],
    )
    for name in [
        "java_jre",
        "javac",
        "python",
        "node_runtime",
        "node_transpiler",
        "llvm_runtime",
        "image_tar",
    ]
]

toolchain(
    name = "java_jre_toolchain",
    toolchain = ":java_jre",
    toolchain_type = "@bazel_tools//tools/jdk:runtime_toolchain_type",
)

toolchain(
    name = "dual_jdk_runtime_toolchain",
    toolchain = ":javac",
    toolchain_type = ":dual_runtime_type",
)

toolchain(
    name = "mixed_build_toolchain",
    toolchain = ":mixed_build",
    toolchain_type = ":mixed_build_type",
)

toolchain(
    name = "mixed_runtime_toolchain",
    toolchain = ":mixed_runtime",
    toolchain_type = ":mixed_runtime_type",
)

toolchain(
    name = "mixed_embedded_toolchain",
    toolchain = ":mixed_embedded",
    toolchain_type = ":mixed_embedded_type",
)

toolchain(
    name = "javac_toolchain",
    toolchain = ":javac",
    toolchain_type = "@bazel_tools//tools/jdk:toolchain_type",
)

toolchain(
    name = "python_toolchain",
    toolchain = ":python",
    toolchain_type = "@bazel_tools//tools/python:toolchain_type",
)

toolchain(
    name = "node_runtime_toolchain",
    toolchain = ":node_runtime",
    toolchain_type = ":node_runtime_type",
)

toolchain(
    name = "node_transpiler_toolchain",
    toolchain = ":node_transpiler",
    toolchain_type = ":node_transpile_type",
)

toolchain(
    name = "llvm_runtime_toolchain",
    toolchain = ":llvm_runtime",
    toolchain_type = ":llvm_toolchain_type",
)

toolchain(
    name = "image_tar_toolchain",
    toolchain = ":image_tar",
    toolchain_type = ":tar_toolchain_type",
)
