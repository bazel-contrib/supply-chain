load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@rules_cc//cc:cc_library.bzl", "cc_library")
load("@supply_chain_tools//sbom:supply_chain.bzl", "supply_chain")

package(
    default_package_metadata = [":package_metadata"],
    default_visibility = ["//visibility:public"],
)

package_metadata(
    name = "package_metadata",
    attributes = [":license"],
    purl = "pkg:generic/supply-chain-example@0.0.1",
)

license(
    name = "license",
    kind = "@package_metadata//licenses/spdx:Apache-2.0",
    text = "LICENSE",
)

cc_library(
    name = "app",
    srcs = ["main.cc"],
)

exports_files(["license-policy.json"])

supply_chain(
    name = "license_compliance",
    target = ":app",
    policy = ":license-policy.json",
    sbom_formats = ["spdx-json", "cyclonedx-json"],
)
