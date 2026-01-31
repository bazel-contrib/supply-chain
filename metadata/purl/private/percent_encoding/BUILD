load("//purl/private/percent_encoding:tables_test.bzl", "test_cases")
load("//purl/private/percent_encoding:test.bzl", "percent_encoding_test")

exports_files(
    [
        "percent_encoding.bzl",
        "tables_test.bzl",
        "tables.bzl",
        "test.bzl",
    ],
    visibility = [
        "//purl/private:__pkg__",
    ],
)

filegroup(
    name = "srcs",
    srcs = [
        "percent_encoding.bzl",
        "tables.bzl",
        "tables_test.bzl",
        "test.bzl",
    ],
    visibility = [
        "//purl/private:__pkg__",
    ],
)

percent_encoding_test(
    name = "percent_encoding_test",
    cases = test_cases,
)
