load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

stardoc(
    name = "license",
    out = "license.generated.md",
    input = "@package_metadata//licenses/rules:license.bzl",
    deps = [
        "@package_metadata//licenses/rules:srcs",
    ],
)

diff_test(
    name = "license_test",
    file1 = ":license",
    file2 = "license.md",
)

stardoc(
    name = "license_kind",
    out = "license_kind.generated.md",
    input = "@package_metadata//licenses/rules:license_kind.bzl",
    deps = [
        "@package_metadata//licenses/rules:srcs",
    ],
)

diff_test(
    name = "license_kind_test",
    file1 = ":license_kind",
    file2 = "license_kind.md",
)
