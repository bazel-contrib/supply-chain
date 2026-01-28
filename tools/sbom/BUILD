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
        ":sbom.bzl",
        ":spdx.bzl",
        "//gather_metadata",
        "//:core_metadata",
    ],
    visibility = ["//visibility:public"],
)
