# Examples of applications and interactions with licenses

load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@rules_cc//cc:cc_binary.bzl", "cc_binary")
load("//sample_reports:licenses_used.bzl", "licenses_used")
load("//sbom:sbom.bzl", "sbom")
load("//vendor/constant_gen:defs.bzl", "constant_gen")

package(
    default_applicable_licenses = ["//:package_metadata"],
    default_visibility = ["//visibility:public"],
)

# Sample
constant_gen(
    name = "message",
    text = "Hello, world.",
    var = "server_message",
)

# This server is clean to ship.
cc_binary(
    name = "my_server",
    srcs = ["server.cc"],
    deps = [":message"],
)

# Verify the licenses are what we expect. The output shows that
# :my_server only uses the unencumbered license type.
licenses_used(
    name = "server_licenses",
    out = "server_licenses.txt",
    target = ":my_server",
)

# This server uses something under a restricted license.
cc_binary(
    name = "my_violating_server",
    srcs = ["server.cc"],
    deps = [
        ":message",
        "//lib:lib_with_third_party_deps",
        "//lib:private_lib",
        "//vendor/acme",
        "//vendor/legacy_license:libfoo",
        "//vendor/libhhgttg",
    ],
)

licenses_used(
    name = "violating_server_licenses",
    out = "violating_server_licenses.txt",
    target = ":my_violating_server",
)

sbom(
    name = "server_sbom",
    out = "server_sbom.txt",
    target = ":my_violating_server",
)

diff_test(
    name = "sbom_diff_test",
    file1 = ":server_sbom.txt",
    file2 = "server_sbom.golden",
)
