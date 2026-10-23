# A package with a commercial license.

load("@package_metadata//:defs.bzl", "package_metadata")
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@rules_cc//cc:cc_library.bzl", "cc_library")

package(
    default_applicable_licenses = [":package_data"],
    default_visibility = ["//visibility:public"],
)

package_metadata(
    name = "package_data",
    attributes = [
        ":license",
    ],
    purl = "pkg:github/acmecorp/coyote@3.1.4?download_url=https://github.com/acmecorp/coyote/archive/refs/tags/3.1.4.tar.gz",
)

# The default license for an entire package is typically named "license".
license(
    name = "license",
    kind = "//my_org/licenses:acme_corp_paid",
    text = "ACME_LICENSE",
)

cc_library(
    name = "acme",
    srcs = ["coyote.cc"],
)
