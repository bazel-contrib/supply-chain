# A package under a single custom license.

load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@rules_cc//cc:cc_library.bzl", "cc_library")

package(
    default_package_metadata = [":package_metadata"],
    default_visibility = ["//visibility:public"],
)

package_metadata(
    name = "package_metadata",
    attributes = [
        ":license",
    ],
    purl = purl.bazel(
        module_name(),
        module_version(),
    ),
    visibility = ["//visibility:public"],
)

license(
    name = "license",
    kind = ":acme_commercial",
    text = "LICENSE",
)

# A custom (non-SPDX) license, declared in the module that owns it so that
# other modules can reuse it.
license_kind(
    name = "acme_commercial",
    full_name = "ACME Commercial License",
    identifier = "LicenseRef-acme-commercial",
    visibility = ["//visibility:public"],
)

cc_library(
    name = "acme_sdk",
)
