load("@bazel_skylib//:bzl_library.bzl", "bzl_library")
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package(
    default_package_metadata = [":package_metadata"],
    default_visibility = ["//visibility:public"],
)

package_metadata(
    name = "package_metadata",
    attributes = [
        ":license",
    ],
    purl = purl.bazel(
        module_name(),
        module_version(),
    ),
    visibility = ["//visibility:public"],
)

# TODO: Investigate why the example/srv targets are not comming up under
# this license. It is as if we get the package_metadata target, but
# then we do not descend to it during aspect traversal.
license(
    name = "license",
    kind = "@package_metadata//licenses/spdx:Apache-2.0",
    text = "LICENSE",
)

filegroup(
    name = "srcs",
    srcs = [
        "//gather_metadata:srcs",
    ],
    visibility = ["//visibility:public"],
)

# This is a convenience warpper to turn the metadata core into a bzl_library
# that can be consumed by others.
bzl_library(
    name = "core_metadata",
    srcs = [
        "@package_metadata//:srcs",
        "@package_metadata//licenses:srcs",
    ],
    visibility = ["//visibility:public"],
)
