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
        "//licenses/private:srcs",
        "//licenses/providers:srcs",
        "//licenses/spdx:licenses.bzl",
        "//providers:srcs",
    ],
    visibility = ["//visibility:public"],
)
