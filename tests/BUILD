load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("//:license_expression_test.bzl", "license_expression_test")
load("//:license_json_test.bzl", "license_json_test")
load("//:license_kind_test.bzl", "license_kind_test")
load("//:purl_test.bzl", "purl_bazel_registry_test")

license_expression_test(
    name = "license_expression_test",
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

purl_bazel_registry_test(
    name = "purl_bazel_registry_test",
)

license_kind_test(
    name = "license_kind_info_plain_test",
    kind = "@package_metadata//licenses/spdx:MIT",
)

license_kind_test(
    name = "license_kind_info_deprecated_test",
    deprecated = True,
    kind = "@package_metadata//licenses/spdx:GPL-2.0",
)

license_kind_test(
    name = "license_kind_info_exception_test",
    exception = True,
    kind = "@package_metadata//licenses/spdx:Classpath-exception-2.0",
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
