"""Rules for generating a SBOM out of a target."""

load("@bazel_skylib//:bzl_library.bzl", "bzl_library")
load("@bazel_skylib//rules:common_settings.bzl", "bool_flag")

package(
    default_package_metadata = ["//:package_metadata"],
    default_visibility = ["//visibility:public"],
)

filegroup(
    name = "srcs",
    srcs = glob(["**"]),
)

# Disabled by default: strict validation runs the upstream CycloneDX CLI or
# SPDX tools-java validator against generated reports, which has a real cost
# across a big dependency graph. Opt in per-invocation with:
#   bazel build //some:target --//sbom:strict_validations
# or flip the default for a whole build with --//sbom:strict_validations=true
# in a bazelrc config.
bool_flag(
    name = "strict_validations",
    build_setting_default = False,
)

toolchain_type(name = "cyclonedx_validator_toolchain_type")

toolchain_type(name = "spdx_validator_toolchain_type")

bzl_library(
    name = "sbom",
    srcs = [
        ":cyclonedx.bzl",
        ":extensions.bzl",
        ":sbom.bzl",
        ":spdx.bzl",
        ":validators.bzl",
        "//:core_metadata",
        "//gather_metadata",
    ],
    visibility = ["//visibility:public"],
)
