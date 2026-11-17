load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//licenses/rules:license_exception.bzl", "license_exception")
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("//:license_exception_test.bzl", "license_exception_test")
load("//:license_expression_analysis_test.bzl", "license_expression_case_test", "license_expression_custom_kinds_test", "license_expression_failure_test", "license_expression_graph_test", "license_expression_or_later_test", "license_expression_parsing_tests")
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

# Not on the SPDX list, and without the LicenseRef- prefix.
license_kind(
    name = "unprefixed_kind",
    full_name = "ACME License",
    identifier = "acme",
    tags = ["manual"],
)

license_kind(
    name = "acme_kind",
    full_name = "ACME",
    identifier = "LicenseRef-acme",
)

license(
    name = "custom_license",
    kind = ":unprefixed_kind",
    tags = ["manual"],
)

license_json_test(
    name = "license_json_custom_test",
    expected = "testdata/custom.license.json",
    license = ":custom_license",
)

license(
    name = "expression_license",
    expression = "MIT OR (Apache-2.0 AND GPL-2.0-only WITH Classpath-exception-2.0) OR LicenseRef-acme",
    kinds = [":acme_kind"],
    tags = ["manual"],
)

license_expression_parsing_tests(
    name = "license_expression_parsing_tests",
    kinds = [
        ":acme_kind",
        ":custom_exception",
    ],
)

license_expression_graph_test(
    name = "license_expression_graph_test",
    expression = "MIT OR GPL-2.0-only WITH Classpath-exception-2.0",
)

license_expression_failure_test(
    name = "license_expression_with_non_exception_test",
    expression = "MIT WITH Apache-2.0",
    message = "'Apache-2.0' is not a license exception and cannot follow 'WITH'",
)

license_expression_failure_test(
    name = "license_expression_exception_as_license_test",
    expression = "Classpath-exception-2.0 OR MIT",
    message = "'Classpath-exception-2.0' is a license exception and can only follow 'WITH'",
)

license_expression_failure_test(
    name = "license_expression_missing_kind_test",
    expression = "MIT OR LicenseRef-unknown",
    message = "no license kind for 'LicenseRef-unknown'",
)

license_expression_failure_test(
    name = "license_expression_unknown_spdx_id_test",
    expression = "MIT OR Foo-1.0",
    message = "no license kind for 'Foo-1.0'",
)

license_exception(
    name = "custom_exception",
    full_name = "ACME linking exception",
    identifier = "AdditionRef-acme-linking",
)

license_expression_custom_kinds_test(
    name = "license_expression_custom_kinds_test",
    expression = "MIT OR (LicenseRef-acme WITH AdditionRef-acme-linking)",
    kinds = [
        ":acme_kind",
        ":custom_exception",
    ],
)

license_expression_custom_kinds_test(
    name = "license_expression_references_case_insensitive_test",
    expression = "MIT OR (LicenseRef-ACME WITH AdditionRef-ACME-Linking)",
    kinds = [
        ":acme_kind",
        ":custom_exception",
    ],
)

license_expression_or_later_test(
    name = "license_expression_or_later_test",
    expression = "GPL-3.0+ OR LGPL-3.0-or-later OR CDDL-1.0+",
)

license_expression_case_test(
    name = "license_expression_case_test",
    expression = "mit OR gpl-2.0-ONLY WITH classpath-exception-2.0",
)

# All three operators, nested: AND inside OR, OR inside AND, WITH inside OR.
license(
    name = "all_operators_license",
    expression = "MIT OR (Apache-2.0 AND (LicenseRef-acme OR GPL-2.0-only WITH Classpath-exception-2.0)) OR (BSD-3-Clause AND LicenseRef-acme WITH AdditionRef-acme-linking)",
    kinds = [
        ":acme_kind",
        ":custom_exception",
    ],
    tags = ["manual"],
)

license_json_test(
    name = "license_json_all_operators_test",
    expected = "testdata/all_operators.license.json",
    license = ":all_operators_license",
)

license(
    name = "or_later_license",
    expression = "GPL-3.0+",
    tags = ["manual"],
)

license_json_test(
    name = "license_json_or_later_test",
    expected = "testdata/or_later.license.json",
    license = ":or_later_license",
)
