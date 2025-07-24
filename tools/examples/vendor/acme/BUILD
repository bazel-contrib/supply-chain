# A package with a commercial license.

load("@package_metadata//licenses/rules:license.bzl", "license")

package(
   default_package_metadata = [":license"],
   default_visibility = ["//examples:__subpackages__"],
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
