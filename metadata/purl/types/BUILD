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
    ],
    visibility = ["//purl:__pkg__"],
)
