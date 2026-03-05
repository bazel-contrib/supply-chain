load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

stardoc(
    name = "defs",
    out = "defs.generated.md",
    input = "@package_metadata_extensions//:defs.bzl",
    deps = [
        "@package_metadata_extensions//:srcs",
    ],
)

diff_test(
    name = "defs_test",
    file1 = ":defs",
    file2 = "defs.md",
)
