# A package with a commercial license.

load("@package_metadata//:defs.bzl", "package_metadata")
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@rules_cc//cc:cc_library.bzl", "cc_library")

package(
    default_package_metadata = [":package_data"],
    default_visibility = ["//examples:__subpackages__"],
)

package_metadata(
    name = "package_data",
    attributes = [
        ":license",
    ],
    purl = "...",
)

# The default license for an entire package is typically named "license".
license(
    name = "license",
    kind = "@supply_chain_tools//examples/my_org/licenses:acme_corp_paid",
    text = "ACME_LICENSE",
)

cc_library(
    name = "acme",
    srcs = ["coyote.cc"],
)
