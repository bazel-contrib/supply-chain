load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

stardoc(
    name = "package_linking_info",
    out = "package_linking_info.generated.md",
    input = "@package_metadata_extensions//providers:package_linking_info.bzl",
    deps = [
        "@package_metadata_extensions//providers:srcs",
    ],
)

diff_test(
    name = "package_linking_info_test",
    file1 = ":package_linking_info",
    file2 = "package_linking_info.md",
)
