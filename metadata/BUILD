load("//licenses/rules:license.bzl", "license")
load("//rules:package_metadata.bzl", "package_metadata")

exports_files(
    [
        "defs.bzl",
    ],
    visibility = ["//visibility:public"],
)

filegroup(
    name = "srcs",
    srcs = [
        "defs.bzl",
    ] + [
        "//providers:srcs",
        "//rules:srcs",
    ],
    visibility = ["//visibility:public"],
)

package_metadata(
    name = "package_metadata",
    attributes = [
        ":license",
    ],
    purl = "pkg:bazel/{}@{}".format(
        module_name(),
        module_version(),
    ),
    visibility = ["//visibility:public"],
)

license(
    name = "license",
    kind = "//licenses/spdx:Apache-2.0",
    text = "LICENSE",
)
