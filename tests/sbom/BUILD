load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@supply_chain_tools//sbom:cyclonedx.bzl", "cyclonedx")
load("@supply_chain_tools//sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//sbom:spdx.bzl", "spdx")
load(":golden_test.bzl", "sbom_golden_test")
load(":strict_validations_test.bzl", "strict_validations_test_suite")

license(
    name = "app_license",
    kind = "@package_metadata//licenses/spdx:Apache-2.0",
)

package_metadata(
    name = "app_metadata",
    attributes = [":app_license"],
    purl = "pkg:generic/app@1.0.0",
)

# Not on the SPDX list: reported by name.
license_kind(
    name = "acme_kind",
    full_name = "ACME License",
    identifier = "LicenseRef-acme",
)

license(
    name = "custom_license",
    kind = ":acme_kind",
)

package_metadata(
    name = "custom_metadata",
    attributes = [":custom_license"],
    purl = "pkg:generic/custom@1.0.0",
)

package_metadata(
    name = "unlicensed_metadata",
    purl = "pkg:generic/unlicensed@1.0.0",
)

filegroup(
    name = "app",
    srcs = [
        ":custom",
        ":unlicensed",
    ],
    applicable_licenses = [":app_metadata"],
)

filegroup(
    name = "custom",
    applicable_licenses = [":custom_metadata"],
)

filegroup(
    name = "unlicensed",
    applicable_licenses = [":unlicensed_metadata"],
)

sbom(
    name = "app_sbom",
    target = ":app",
)

spdx(
    name = "app_spdx",
    sbom = ":app_sbom",
    format = "json-ld",
)

spdx(
    name = "app_spdx23",
    sbom = ":app_sbom",
)

cyclonedx(
    name = "app_cyclonedx",
    sbom = ":app_sbom",
)

sbom_golden_test(
    name = "app_spdx_golden_test",
    golden = "testdata/app_spdx.json",
    report = ":app_spdx",
)

sbom_golden_test(
    name = "app_spdx23_golden_test",
    golden = "testdata/app_spdx23.json",
    report = ":app_spdx23",
)

sbom_golden_test(
    name = "app_cyclonedx_golden_test",
    golden = "testdata/app_cyclonedx.json",
    report = ":app_cyclonedx",
)

strict_validations_test_suite(
    name = "strict_validations_tests",
    cyclonedx = ":app_cyclonedx",
    spdx = ":app_spdx",
)
