exports_files(
    [
        "license.bzl",
        "license_kind.bzl",
        "license_exception.bzl",
    ],
    visibility = ["//visibility:public"],
)

filegroup(
    name = "srcs",
    srcs = [
        "license.bzl",
        "license_exception.bzl",
        "license_kind.bzl",
    ] + [
        "//licenses/providers:srcs",
        "//providers:srcs",
    ],
    visibility = ["//visibility:public"],
)
