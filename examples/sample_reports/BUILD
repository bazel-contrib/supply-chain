# Examples of metadata reporting tools
#

load("@supply_chain_tools//sbom:cyclonedx.bzl", "cyclonedx")
load("@supply_chain_tools//sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//sbom:spdx.bzl", "spdx")

package(default_visibility = ["//visibility:public"])

sbom(
    name = "library_sbom",
    target = "//lib:lib_with_third_party_deps",
)

cyclonedx(
    name = "library_cyclonedx",
    sbom = ":library_sbom",
)

spdx(
    name = "library_spdx",
    sbom = ":library_sbom",
)
