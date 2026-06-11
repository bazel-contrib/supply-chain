exports_files(
    [
        "builder.bzl",
        "parser.bzl",
    ],
    visibility = [
        "//:__subpackages__",
    ],
)

filegroup(
    name = "srcs",
    srcs = [
        "builder.bzl",
        "parser.bzl",
    ] + [
        "//private/normalization:srcs",
        "//private/percent_encoding:srcs",
        "//private/strings:srcs",
        "//private/validation:srcs",
    ],
    visibility = [
        "//:__subpackages__",
    ],
)
