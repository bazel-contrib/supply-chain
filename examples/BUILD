load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

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
    kind = "@package_metadata//licenses/spdx:Apache-2.0",
    text = "LICENSE",
)

# A package available under a choice of licenses, parsed from an SPDX license
# expression. This is how https://github.com/serde-rs/serde is licensed.
license(
    name = "dual_license",
    expression = "Apache-2.0 OR MIT",
)

# `WITH` attaches an SPDX license exception to a license. Here the package is
# available under either MIT, or Apache-2.0 combined with the LLVM exception.
license(
    name = "dual_license_with_exception",
    expression = "MIT OR (Apache-2.0 WITH LLVM-exception)",
)

# Custom licenses and exceptions are declared as `license_kind`s (see
# //my_org/licenses) and passed in `kinds`, since only SPDX identifiers are
# resolved automatically. This reads: either MIT, or the ACME commercial
# license with the ACME linking exception.
license(
    name = "custom_license_expression",
    expression = "MIT OR (LicenseRef-acme-commercial WITH AdditionRef-acme-linking)",
    kinds = [
        "//my_org/licenses:acme_commercial",
        "//my_org/licenses:acme_linking_exception",
    ],
)
