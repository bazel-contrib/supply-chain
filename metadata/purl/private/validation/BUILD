exports_files(
    [
        "helpers.bzl",
    ],
    visibility = ["//visibility:public"],
)

exports_files(
    [
        "validation.bzl",
    ],
    visibility = [
        "//purl/private:__pkg__",
        "//purl/types:__pkg__",
    ],
)

filegroup(
    name = "srcs",
    srcs = [
        "helpers.bzl",
        "validation.bzl",
    ],
    visibility = [
        "//purl/private:__pkg__",
        "//purl/types:__pkg__",
    ],
)
