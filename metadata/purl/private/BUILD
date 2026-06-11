exports_files(
    [
        "builder.bzl",
        "parser.bzl",
    ],
    visibility = [
        "//purl:__subpackages__",
    ],
)

filegroup(
    name = "srcs",
    srcs = [
        "builder.bzl",
        "parser.bzl",
    ] + [
        "//purl/private/normalization:srcs",
        "//purl/private/percent_encoding:srcs",
        "//purl/private/strings:srcs",
        "//purl/private/validation:srcs",
    ],
    visibility = [
        "//purl:__subpackages__",
    ],
)
