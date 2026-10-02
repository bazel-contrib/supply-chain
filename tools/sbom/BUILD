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

# Disabled by default: schema validation re-parses and re-validates every
# generated SBOM against the (large) vendored CycloneDX/SPDX JSON Schemas on
# every build, which has a real cost across a big dependency graph. Opt in
# per-invocation with:
#   bazel build //some:target --//sbom:validate_schemas
# or flip the default for a whole build with --//sbom:validate_schemas=true
# in a bazelrc config.
bool_flag(
    name = "validate_schemas",
    build_setting_default = False,
)

bzl_library(
    name = "sbom",
    srcs = [
        ":cyclonedx.bzl",
        ":sbom.bzl",
        ":spdx.bzl",
        "//:core_metadata",
        "//gather_metadata",
    ],
    visibility = ["//visibility:public"],
)
