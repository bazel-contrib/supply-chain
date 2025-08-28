load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

stardoc(
    name = "gather_metadata",
    out = "gather_metadata.generated.md",
    input = "@supply_chain_tools//gather_metadata:gather_metadata.bzl",
    deps = [
        "@supply_chain_tools//gather_metadata",
    ],
)

diff_test(
    name = "gather_metadata_test",
    file1 = ":gather_metadata.generated.md",
    file2 = "gather_metadata.md",
)
