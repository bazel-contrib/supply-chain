"""Rules for collecting package_metadata."""

load("@bazel_skylib//:bzl_library.bzl", "bzl_library")
load("@bazel_skylib//lib:unittest.bzl", "unittest")
load(":attr_types_test.bzl", "attr_types_test_suite")
load(":dedup_test.bzl", "dedup_test_suite")
load(":filtering_test.bzl", "filtering_test_suite")
load(":integration_test.bzl", "integration_test_suite")
load(":serialization_integration_test.bzl", "serialization_integration_test_suite")
load(":toolchain_involvement_test.bzl", "toolchain_involvement_test_suite")

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
        ":toolchain_involvement.bzl",
        "//:core_metadata",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@bazel_features//:features",
        "@bazel_skylib//rules:common_settings",
    ],
)

bzl_library(
    name = "toolchain_involvement",
    srcs = [":toolchain_involvement.bzl"],
    visibility = ["//visibility:public"],
)

bzl_library(
    name = "extensions",
    srcs = [":extensions.bzl"],
    visibility = ["//visibility:public"],
    deps = [":toolchain_involvement"],
)

integration_test_suite(name = "integration_tests")

serialization_integration_test_suite(name = "serialization_integration_tests")

filtering_test_suite(name = "filtering_tests")

attr_types_test_suite(name = "attr_types_tests")

dedup_test_suite(name = "dedup_tests")

toolchain_involvement_test_suite(name = "toolchain_involvement_tests")

test_suite(
    name = "all_tests",
    tests = [
        ":attr_types_tests",
        ":dedup_tests",
        ":filtering_tests",
        ":integration_tests",
        ":serialization_integration_tests",
        ":toolchain_involvement_tests",
    ],
)
