# A package under a choice between a standard license with a standard exception, and a custom license with a custom exception.

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
    expression = "(Apache-2.0 WITH LLVM-exception) OR (LicenseRef-acme-commercial WITH AdditionRef-acme-linking)",
    kinds = [
        ":acme_linking_exception",
        "@acme_sdk//:acme_commercial",
    ],
    text = "LICENSE",
)

# A custom license exception. Custom exceptions use the `AdditionRef-` prefix
# and must set `exception = True`.
license_kind(
    name = "acme_linking_exception",
    exception = True,
    full_name = "ACME Linking Exception",
    identifier = "AdditionRef-acme-linking",
    visibility = ["//visibility:public"],
)

cc_library(
    name = "acme_plugin",
    deps = ["@acme_sdk"],
)
