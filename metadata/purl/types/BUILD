load("//purl/private:builtin_type.bzl", "builtin_type")
load("//purl:type_names.bzl", "BUILTIN_PURL_TYPES")

[
    builtin_type(
        name = type,
        visibility = ["//visibility:public"],
    )
    for type in BUILTIN_PURL_TYPES
]

filegroup(
    name = "srcs",
    srcs = [
        "BUILD",
        "alpm.bzl",
        "apk.bzl",
        "bazel.bzl",
        "bitbucket.bzl",
        "bitnami.bzl",
        "cargo.bzl",
        "chrome_extension.bzl",
        "cocoapods.bzl",
        "composer.bzl",
        "conan.bzl",
        "conda.bzl",
        "cpan.bzl",
        "cran.bzl",
        "deb.bzl",
        "docker.bzl",
        "gem.bzl",
        "generic.bzl",
        "github.bzl",
        "golang.bzl",
        "hackage.bzl",
        "hex.bzl",
        "huggingface.bzl",
        "julia.bzl",
        "luarocks.bzl",
        "maven.bzl",
        "mlflow.bzl",
        "npm.bzl",
        "nuget.bzl",
        "oci.bzl",
        "opam.bzl",
        "otp.bzl",
        "pub.bzl",
        "pypi.bzl",
        "qpkg.bzl",
        "rpm.bzl",
        "swid.bzl",
        "swift.bzl",
        "vscode_extension.bzl",
        "yocto.bzl",
    ],
    visibility = ["//purl:__pkg__"],
)
