load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("//:license_expression_analysis_test.bzl", "license_expression_failure_test", "license_expression_custom_kinds_test", "license_expression_graph_test")
load("//:license_expression_test.bzl", "license_expression_test")
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
    expression = "MIT OR (Apache-2.0 AND GPL-2.0-only+ WITH Classpath-exception-2.0) OR LicenseRef-acme",
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
    expression = "MIT OR GPL-2.0-only+ WITH Classpath-exception-2.0",
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
