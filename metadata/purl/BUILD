exports_files(
    [
        "extensions.bzl",
        "purl.bzl",
        "providers.bzl",
        "type_names.bzl",
    ],
    visibility = ["//visibility:public"],
)

filegroup(
    name = "srcs",
    srcs = [
        "extensions.bzl",
        "purl.bzl",
        "providers.bzl",
        "type_names.bzl",
    ] + [
        "//purl/private:srcs",
        "//purl/types:srcs",
    ],
    visibility = ["//visibility:public"],
)
