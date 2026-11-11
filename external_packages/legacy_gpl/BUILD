# A package using a deprecated SPDX identifier (`GPL-2.0-with-classpath-exception`, now `GPL-2.0-only WITH Classpath-exception-2.0`) and the `+` shorthand, kept as an or-later flag.

load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@rules_cc//cc:cc_library.bzl", "cc_library")

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

license(
    name = "license",
    expression = "GPL-2.0+ OR GPL-2.0-with-classpath-exception",
    text = "LICENSE",
)

cc_library(
    name = "legacy_gpl",
)
