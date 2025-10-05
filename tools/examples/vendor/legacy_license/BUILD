# A package with all code under a single license. This is the most common case
# we expect to see.

load("@rules_license//rules:license.bzl", "license")
load("@rules_cc//cc:cc_library.bzl", "cc_library")

# Using a package wide default ensure that all targets are associated with the
# license.
package(
    default_package_metadata = [":license"],
    default_visibility = ["//examples:__subpackages__"],
)

# The default license for an entire package is typically named "license".
license(
    name = "license",
    license_kinds = ["@rules_license//licenses/generic:reciprocal"],
    license_text = "COPYING",
)

cc_library(
    name = "libfoo",
    srcs = ["foo.cc"],
)
