# A package with a commercial license.

load("@package_metadata//:defs.bzl", "package_metadata")
load("@package_metadata//attributes:originator.bzl", "originator")
load("@package_metadata//attributes:supplier.bzl", "supplier")
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@rules_cc//cc:cc_library.bzl", "cc_library")

package(
    default_package_metadata = [":package_data"],
    default_visibility = ["//visibility:public"],
)

package_metadata(
    name = "package_data",
    attributes = [
        ":license",
        ":supplier",
        ":originator",
    ],
    purl = "pkg:github/acmecorp/coyote@3.1.4?download_url=https://github.com/acmecorp/coyote/archive/refs/tags/3.1.4.tar.gz",
)

supplier(
    name = "supplier",
    email = "oss@acme.example",
    entity_name = "ACME Corp",
    url = "https://acme.example",
)

originator(
    name = "originator",
    email = "coyote@acme.example",
    entity_name = "A. Coyote",
    type = "person",
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
