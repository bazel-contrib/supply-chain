exports_files(
    [
        "license.bzl",
        "license_kind.bzl",
    ],
    visibility = ["//visibility:public"],
)

filegroup(
    name = "srcs",
    srcs = [
        "license.bzl",
        "license_kind.bzl",
    ] + [
        "//licenses/private:srcs",
        "//licenses/providers:srcs",
        "//licenses/spdx:or_later.bzl",
        "//providers:srcs",
    ],
    visibility = ["//visibility:public"],
)
