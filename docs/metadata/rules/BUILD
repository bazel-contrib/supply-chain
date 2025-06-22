load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

stardoc(
    name = "package_metadata",
    out = "package_metadata.generated.md",
    input = "@package_metadata//rules:package_metadata.bzl",
    deps = [
        "@package_metadata//rules:srcs",
    ],
)

diff_test(
    name = "package_metadata_test",
    file1 = ":package_metadata",
    file2 = "package_metadata.md",
)
