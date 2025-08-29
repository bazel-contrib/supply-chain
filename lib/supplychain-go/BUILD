load("@gazelle//:def.bzl", "gazelle")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@rules_go//go:def.bzl", "go_library", "go_test")

package_metadata(
    name = "package_metadata",
    purl = "pkg:bazel/{}@{}".format(
        module_name(),
        module_version(),
    ),
)

gazelle(name = "gazelle")

go_library(
    name = "supplychain-go",
    srcs = [
        "package_attribute.go",
        "package_metadata.go",
    ],
    importpath = "github.com/bazel-contrib/supply-chain/lib/supplychain-go",
    visibility = ["//visibility:public"],
    deps = [
        "//label",
        "@com_github_package_url_packageurl_go//:packageurl-go",
    ],
)

go_test(
    name = "supplychain-go_test",
    srcs = ["package_metadata_test.go"],
    embed = [":supplychain-go"],
    deps = [
        "//label",
        "@com_github_package_url_packageurl_go//:packageurl-go",
        "@com_github_stretchr_testify//assert",
    ],
)
