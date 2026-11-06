load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//licenses/rules:license_exception.bzl", "license_exception")
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("//:license_exception_test.bzl", "license_exception_test")
load("//:license_json_test.bzl", "license_json_test")
load("//:license_kind_test.bzl", "license_kind_test")
load("//:purl_test.bzl", "purl_bazel_registry_test")
load("//:spdx_alias_test.bzl", "spdx_alias_test")

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

purl_bazel_registry_test(
    name = "purl_bazel_registry_test",
)

license_kind_test(
    name = "license_kind_info_plain_test",
    kind = "@package_metadata//licenses/spdx:MIT",
)

license_kind_test(
    name = "license_kind_info_deprecated_test",
    deprecation = "Deprecated since SPDX license list 3.0; use GPL-2.0-only or GPL-2.0-or-later (GPL-2.0+) instead.",
    kind = "@package_metadata//licenses/spdx:GPL-2.0",
)

spdx_alias_test(
    name = "spdx_license_alias_test",
    identifier = "MIT",
    lowercase = "@package_metadata//licenses/spdx:mit",
    official = "@package_metadata//licenses/spdx:MIT",
)

license_exception(
    name = "test_exception",
    full_name = "Test exception",
    identifier = "Test-exception",
)

license_exception_test(
    name = "license_exception_info_test",
    exception = ":test_exception",
    exception_name = "Test exception",
    identifier = "Test-exception",
)

license(
    name = "single_license",
    kind = "@package_metadata//licenses/spdx:MIT",
    tags = ["manual"],
)

license_json_test(
    name = "license_json_single_test",
    expected = "testdata/single.license.json",
    license = ":single_license",
)

license_kind(
    name = "acme_kind",
    full_name = "ACME License",
    identifier = "acme",
    tags = ["manual"],
)

license(
    name = "custom_license",
    kind = ":acme_kind",
    tags = ["manual"],
)

license_json_test(
    name = "license_json_custom_test",
    expected = "testdata/custom.license.json",
    license = ":custom_license",
)
