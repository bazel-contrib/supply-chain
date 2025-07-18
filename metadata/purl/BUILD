load("//purl:percent_encoding_test.bzl", "percent_encoding_test")
load("//purl/private:tables_test.bzl", "test_cases")

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
    cases = test_cases["percent_encoding"]["tests"],
)
