load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

stardoc(
    name = "purl",
    out = "purl.generated.md",
    input = "@package_metadata//purl:purl.bzl",
    deps = [
        "@package_metadata//purl:srcs",
    ],
)

diff_test(
    name = "purl_test",
    file1 = ":purl",
    file2 = "purl.md",
)
