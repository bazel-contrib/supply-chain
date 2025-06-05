load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

stardoc(
    name = "package_attribute_override",
    out = "package_attribute_override.generated.md",
    input = "@package_metadata//rules:package_attribute_override.bzl",
    deps = [
        "@package_metadata//rules:srcs",
    ],
)

diff_test(
    name = "package_attribute_override_test",
    file1 = ":package_attribute_override",
    file2 = "package_attribute_override.md",
)

stardoc(
    name = "package_metadata_toolchain",
    out = "package_metadata_toolchain.generated.md",
    input = "@package_metadata//rules:package_metadata_toolchain.bzl",
    deps = [
        "@package_metadata//rules:srcs",
    ],
)

diff_test(
    name = "package_metadata_toolchain_test",
    file1 = ":package_metadata_toolchain",
    file2 = "package_metadata_toolchain.md",
)

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
