# A package with all code under a single license. This is the most common case
# we expect to see.

load("@package_metadata//licenses/rules:license.bzl", "license")

# Using a package wide default ensure that all targets are associated with the
# license.
package(
    default_package_metadata = [":license"],
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
