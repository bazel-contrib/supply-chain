load("//purl:percent_encoding_test.bzl", "percent_encoding_test")
load("//purl:starlark.bzl", "starlark")
load("//purl:tables.bzl", "percent_encoding")

print(starlark.encoding)

exports_files(
    [
        "purl.bzl",
    ],
    visibility = ["//visibility:public"],
)

filegroup(
    name = "srcs",
    srcs = [
        "purl.bzl",
    ],
    visibility = ["//visibility:public"],
)

percent_encoding_test(
    name = "percent_encoding_test",
    cases = percent_encoding.tests,
)
