exports_files(
    [
        "package_metadata_override.bzl",
    ],
    visibility = ["//visibility:public"],
)

filegroup(
    name = "srcs",
    srcs = [
        "package_group.bzl",
        "package_metadata_override.bzl",
    ] + [
        "//providers:srcs",
        "@package_metadata//providers:srcs",
    ],
    visibility = ["//visibility:public"],
)
