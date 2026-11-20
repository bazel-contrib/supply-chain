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

# True when the build is invoked with --stamp. The `cyclonedx` and `spdx` rules
# use it as the default of their `stamp` attribute.
config_setting(
    name = "stamp",
    values = {"stamp": "1"},
)

# Exec-configuration dependencies remain excluded by default. Opt in when
# build tools should be included and classified as build-only SBOM edges.
bool_flag(
    name = "include_build_tools",
    build_setting_default = False,
)

# Disabled by default: fails analysis of an `sbom` when the package metadata of
# a component does not declare a `component_type` attribute, instead of
# reporting the format's fallback (`library` in CycloneDX, no purpose in SPDX).
# Opt in per-invocation with:
#   bazel build //some:target --//sbom:strict_component_types
bool_flag(
    name = "strict_component_types",
    build_setting_default = False,
)

# Disabled by default: prints a warning listing the components that do not
# declare a `component_type` attribute when generating a CycloneDX report. The
# list can be very long, so it is opt-in:
#   bazel build //some:target --//sbom:verbose_component_types
bool_flag(
    name = "verbose_component_types",
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
