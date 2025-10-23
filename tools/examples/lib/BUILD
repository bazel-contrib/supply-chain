# Examples of applications and interactions with licenses

load("@rules_cc//cc:cc_library.bzl", "cc_library")
load("//examples/sample_reports:licenses_used.bzl", "licenses_used")

package(
    default_visibility = ["//examples:__subpackages__"],
)

cc_library(
    name = "private_lib",
    srcs = [
        "private_lib.cc",
    ],
    hdrs = [
        "private_lib.h",
    ],
)

cc_library(
    name = "lib_with_third_party_deps",
    deps = [
        "//examples/vendor/acme",
    ],
)

# Verify the licenses are what we expect. The output shows that
# :my_server only uses the unencumbered license type.
licenses_used(
    name = "server_licenses",
    out = "server_licenses.txt",
    target = ":lib_with_third_party_deps",
)

#sbom(
#    name = "server_sbom",
#    out = "server_sbom.txt",
#    target = ":lib_with_third_party_deps",
#)
