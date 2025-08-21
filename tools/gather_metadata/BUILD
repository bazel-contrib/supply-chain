"""Rules for collecting package_metadata."""

load("@bazel_skylib//:bzl_library.bzl", "bzl_library")
load(":trace.bzl", "trace")

package(
    default_package_metadata = ["//:package_metadata"],
    default_visibility = ["//visibility:public"],
)

filegroup(
    name = "srcs",
    srcs = glob(["**"]),
)

bzl_library(
    name = "gather_metadata",
    srcs = [
        ":core.bzl",
        ":gather_metadata.bzl",
        ":providers.bzl",
        ":rule_filters.bzl",
        ":trace.bzl",
        "//:core_metadata",
    ],
    visibility = ["//visibility:public"],
)

# This target controls the value of the traced target used during dependency collection.
# It is useful for debugging unexpected dependencies.
# This value should always be the empty string!
# Specify this value with a flag, like --@supply_chain_tools//gather_metadata:trace_target=//target/to:trace
trace(
    name = "trace_target",
    build_setting_default = "",  # TRACE-TARGET-SHOULD-BE-EMPTY
    visibility = ["//visibility:public"],
)
