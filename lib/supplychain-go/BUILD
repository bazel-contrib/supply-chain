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
    name = "supply-chain-go",
    srcs = [
        "package_attribute.go",
        "package_metadata.go",
    ],
    importpath = "github.com/bazel-contrib/supply-chain/lib/supply-chain-go",
    visibility = ["//visibility:public"],
    deps = ["@com_github_package_url_packageurl_go//:packageurl-go"],
)

go_test(
    name = "supply-chain-go_test",
    srcs = ["package_metadata_test.go"],
    embed = [":supply-chain-go"],
    deps = [
        "@com_github_package_url_packageurl_go//:packageurl-go",
        "@com_github_stretchr_testify//assert",
    ],
)
