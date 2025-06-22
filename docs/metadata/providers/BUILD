load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

stardoc(
    name = "package_attribute_info",
    out = "package_attribute_info.generated.md",
    input = "@package_metadata//providers:package_attribute_info.bzl",
    deps = [
        "@package_metadata//providers:srcs",
    ],
)

diff_test(
    name = "package_attribute_info_test",
    file1 = ":package_attribute_info",
    file2 = "package_attribute_info.md",
)

stardoc(
    name = "package_metadata_info",
    out = "package_metadata_info.generated.md",
    input = "@package_metadata//providers:package_metadata_info.bzl",
    deps = [
        "@package_metadata//providers:srcs",
    ],
)

diff_test(
    name = "package_metadata_info_test",
    file1 = ":package_metadata_info",
    file2 = "package_metadata_info.md",
)
