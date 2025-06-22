load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

stardoc(
    name = "license_kind_info",
    out = "license_kind_info.generated.md",
    input = "@package_metadata//licenses/providers:license_kind_info.bzl",
    deps = [
        "@package_metadata//licenses/providers:srcs",
    ],
)

diff_test(
    name = "license_kind_info_test",
    file1 = ":license_kind_info",
    file2 = "license_kind_info.md",
)
