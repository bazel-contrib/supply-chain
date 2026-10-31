load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("//:license_expression_analysis_test.bzl", "license_expression_failure_test", "license_expression_custom_kinds_test", "license_expression_graph_test", "license_expression_or_later_test", "license_expression_case_test")
load("//:license_expression_test.bzl", "license_expression_test")
load("//:license_json_test.bzl", "license_json_test")
load("//:purl_test.bzl", "purl_bazel_registry_test")

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

license_expression_test(
    name = "license_expression_test",
)

license(
    name = "expression_license",
    expression = "MIT OR (Apache-2.0 AND GPL-2.0-only WITH Classpath-exception-2.0) OR LicenseRef-acme",
    kinds = [":acme_kind"],
    tags = ["manual"],
)

license_kind(
    name = "acme_kind",
    full_name = "ACME",
    identifier = "LicenseRef-acme",
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

license_kind(
    name = "custom_exception",
    exception = True,
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

license_expression_failure_test(
    name = "license_expression_custom_exception_not_flagged_test",
    expression = "LicenseRef-acme WITH AdditionRef-unflagged",
    kinds = [
        ":acme_kind",
        ":unflagged_exception",
    ],
    message = "'AdditionRef-unflagged' is not a license exception and cannot follow 'WITH'",
)

license_kind(
    name = "unflagged_exception",
    full_name = "Exception missing the exception flag",
    identifier = "AdditionRef-unflagged",
)

license_expression_or_later_test(
    name = "license_expression_or_later_test",
    expression = "GPL-3.0+ OR LGPL-3.0-or-later OR CDDL-1.0+",
)

license_expression_case_test(
    name = "license_expression_case_test",
    expression = "mit OR gpl-2.0-ONLY WITH classpath-exception-2.0",
)

license_expression_failure_test(
    name = "license_kind_exception_test",
    kind = "@package_metadata//licenses/spdx:Classpath-exception-2.0",
    message = "'Classpath-exception-2.0' is a license exception and can only follow 'WITH'",
)

license_kind(
    name = "invalid_identifier_kind",
    full_name = "Identifier with a character SPDX does not allow",
    identifier = "LicenseRef-acme_internal",
    tags = ["manual"],
)

license_expression_failure_test(
    name = "license_kind_invalid_identifier_test",
    kind = ":invalid_identifier_kind",
    message = "invalid license identifier 'LicenseRef-acme_internal'*found '_'",
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
    name = "single_license",
    kind = "@package_metadata//licenses/spdx:MIT",
    tags = ["manual"],
)

license_json_test(
    name = "license_json_single_test",
    expected = "testdata/single.license.json",
    license = ":single_license",
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
