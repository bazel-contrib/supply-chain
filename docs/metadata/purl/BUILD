load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@bazel_skylib//:bzl_library.bzl", "bzl_library")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

bzl_library(
    name = "type_registries",
    srcs = [
        "@package_metadata//purl/types:requirements.bzl",
        "@package_metadata//purl/types:helpers.bzl",
        "@package_metadata_purl_types//:normalizers.bzl",
        "@package_metadata_purl_types//:validators.bzl",
    ],
    visibility = ["//visibility:public"],
)

stardoc(
    name = "purl",
    out = "purl.generated.md",
    input = "@package_metadata//purl:purl.bzl",
    deps = [
        "@package_metadata//purl:srcs",
        ":type_registries",
    ],
)

diff_test(
    name = "purl_test",
    file1 = ":purl",
    file2 = "purl.md",
)
