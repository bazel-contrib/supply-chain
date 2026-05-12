"""Rules for collecting package_metadata."""

load("@bazel_skylib//:bzl_library.bzl", "bzl_library")
load("@bazel_skylib//lib:unittest.bzl", "unittest")
load(":integration_test.bzl", "integration_test_suite")
load(":serialization_test.bzl", "serialization_test_suite")

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
        ":serialization.bzl",
        "//:core_metadata",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@bazel_features//:features",
    ],
)

# Unit tests for serialization logic
serialization_test_suite(name = "serialization_tests")

# Integration tests for end-to-end workflows
integration_test_suite(name = "integration_tests")

# Combined test suite
test_suite(
    name = "all_tests",
    tests = [
        ":serialization_tests",
        ":integration_tests",
    ],
)
