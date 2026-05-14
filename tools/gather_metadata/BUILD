"""Rules for collecting package_metadata."""

load("@bazel_skylib//:bzl_library.bzl", "bzl_library")
load("@bazel_skylib//lib:unittest.bzl", "unittest")
load(":integration_test.bzl", "integration_test_suite")
load(":serialization_integration_test.bzl", "serialization_integration_test_suite")
load(":filtering_test.bzl", "filtering_test_suite")

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

integration_test_suite(name = "integration_tests")

serialization_integration_test_suite(name = "serialization_integration_tests")

filtering_test_suite(name = "filtering_tests")

test_suite(
    name = "all_tests",
    tests = [
        ":integration_tests",
        ":serialization_integration_tests",
        ":filtering_tests",
    ],
)
