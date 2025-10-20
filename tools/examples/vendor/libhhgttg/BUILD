# A package with all code under a single license. This is the most common case
# we expect to see.

load("@package_metadata//licenses/rules:license.bzl", "license")
load("@rules_cc//cc:cc_library.bzl", "cc_library")

# Using a package wide default ensure that all targets are associated with the
# license.
package(
    default_applicable_licenses = [":license"],
    default_visibility = ["//examples:__subpackages__"],
)

# The default license for an entire package is typically named "license".
license(
    name = "license",
    kind = "@supply_chain_tools//examples/my_org/licenses:generic_notice",
    text = "LICENSE",
)

cc_library(
    name = "libhhgttg",
    srcs = ["answer.cc"],
)
