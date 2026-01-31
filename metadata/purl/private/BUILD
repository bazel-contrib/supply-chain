exports_files(
    [
        "builder.bzl",
    ],
    visibility = [
        "//purl:__subpackages__",
    ],
)

filegroup(
    name = "srcs",
    srcs = [
        "builder.bzl",
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
