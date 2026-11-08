load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@supply_chain_tools//sbom:cyclonedx.bzl", "cyclonedx")
load("@supply_chain_tools//sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//sbom:spdx.bzl", "spdx")
load(":golden_test.bzl", "sbom_golden_test")
load(":strict_validations_test.bzl", "strict_validations_test_suite")

package_metadata(
    name = "app_metadata",
    purl = "pkg:generic/app@1.0.0",
)

package_metadata(
    name = "dep_metadata",
    purl = "pkg:generic/dep@2.0.0",
)

filegroup(
    name = "app",
    srcs = [":dep"],
    applicable_licenses = [":app_metadata"],
)

filegroup(
    name = "dep",
    applicable_licenses = [":dep_metadata"],
)

sbom(
    name = "app_sbom",
    target = ":app",
)

spdx(
    name = "app_spdx",
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
    name = "app_cyclonedx_golden_test",
    golden = "testdata/app_cyclonedx.json",
    report = ":app_cyclonedx",
)

strict_validations_test_suite(
    name = "strict_validations_tests",
    cyclonedx = ":app_cyclonedx",
    spdx = ":app_spdx",
)
