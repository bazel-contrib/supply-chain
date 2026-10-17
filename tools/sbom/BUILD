"""Rules for generating a SBOM out of a target."""

load("@bazel_skylib//:bzl_library.bzl", "bzl_library")

package(
    default_package_metadata = ["//:package_metadata"],
    default_visibility = ["//visibility:public"],
)

filegroup(
    name = "srcs",
    srcs = glob(["**"]),
)

bzl_library(
    name = "sbom",
    srcs = [
        ":cyclonedx.bzl",
        ":license_coverage.bzl",
        ":license_policy.bzl",
        ":sbom.bzl",
        ":spdx.bzl",
        ":supply_chain.bzl",
        "//gather_metadata",
        "//:core_metadata",
    ],
    visibility = ["//visibility:public"],
)
