load("@bazel_skylib//rules:build_test.bzl", "build_test")
load("@bazel_skylib//rules:write_file.bzl", "write_file")
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//licenses/rules:license_exception.bzl", "license_exception")
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@supply_chain_tools//sbom:cyclonedx.bzl", "cyclonedx")
load("@supply_chain_tools//sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//sbom:spdx.bzl", "spdx")
load(":expected_failure.bzl", "cyclonedx_expected_failure", "spdx_expected_failure")
load(":golden_test.bzl", "sbom_golden_test", "sbom_timestamp_test")
load(":strict_component_types_test.bzl", "strict_component_types_test_suite")
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

license_exception(
    name = "acme_linking_kind",
    full_name = "ACME linking exception",
    identifier = "AdditionRef-acme-linking",
)

license(
    name = "choice_license",
    expression = "MIT OR (Apache-2.0 AND LicenseRef-acme)",
    kinds = [":acme_kind"],
)

package_metadata(
    name = "choice_metadata",
    attributes = [":choice_license"],
    purl = "pkg:generic/choice@1.0.0",
)

license(
    name = "exception_license",
    expression = "GPL-2.0-only WITH Classpath-exception-2.0 OR LicenseRef-acme WITH AdditionRef-acme-linking",
    kinds = [
        ":acme_kind",
        ":acme_linking_kind",
    ],
)

package_metadata(
    name = "exception_metadata",
    attributes = [":exception_license"],
    purl = "pkg:generic/with_exception@1.0.0",
)

license(
    name = "or_later_license",
    expression = "GPL-3.0+",
)

package_metadata(
    name = "or_later_metadata",
    attributes = [":or_later_license"],
    purl = "pkg:generic/or_later@1.0.0",
)

package_metadata(
    name = "unlicensed_metadata",
    purl = "pkg:generic/unlicensed@1.0.0",
)

filegroup(
    name = "app",
    srcs = [
        ":choice",
        ":custom",
        ":or_later",
        ":unlicensed",
        ":with_exception",
    ],
    applicable_licenses = [":app_metadata"],
)

filegroup(
    name = "choice",
    applicable_licenses = [":choice_metadata"],
)

filegroup(
    name = "with_exception",
    applicable_licenses = [":exception_metadata"],
)

filegroup(
    name = "or_later",
    applicable_licenses = [":or_later_metadata"],
)

filegroup(
    name = "custom",
    applicable_licenses = [":custom_metadata"],
)

filegroup(
    name = "unlicensed",
    applicable_licenses = [":unlicensed_metadata"],
)

# SPDX 2.3 cannot represent a custom exception, so it has its own graph, with
# an exception of the SPDX list only.
license(
    name = "listed_exception_license",
    expression = "GPL-2.0-only WITH Classpath-exception-2.0 OR LicenseRef-acme",
    kinds = [":acme_kind"],
)

package_metadata(
    name = "listed_exception_metadata",
    attributes = [":listed_exception_license"],
    purl = "pkg:generic/with_listed_exception@1.0.0",
)

filegroup(
    name = "with_listed_exception",
    applicable_licenses = [":listed_exception_metadata"],
)

filegroup(
    name = "app23",
    srcs = [
        ":choice",
        ":custom",
        ":or_later",
        ":unlicensed",
        ":with_listed_exception",
    ],
    applicable_licenses = [":app_metadata"],
)

sbom(
    name = "app_sbom",
    target = ":app",
)

sbom(
    name = "app23_sbom",
    target = ":app23",
)

spdx(
    name = "app_spdx",
    format = "json-ld",
    sbom = ":app_sbom",
)

spdx(
    name = "app_spdx23",
    sbom = ":app23_sbom",
)

cyclonedx(
    name = "app_cyclonedx",
    sbom = ":app_sbom",
)

# The serial number and the document namespace of a report come from a file, so
# that a build can produce them, for example from its workspace status.
write_file(
    name = "serial_number",
    out = "serial_number.txt",
    content = ["urn:uuid:7c9e6679-7425-40de-944b-e07fc1f90ae7"],
)

write_file(
    name = "document_namespace",
    out = "document_namespace.txt",
    content = ["https://example.com/spdxdocs/app-1.0.0-4d3c2b1a"],
)

cyclonedx(
    name = "app_cyclonedx_serial_number",
    sbom = ":app_sbom",
    serial_number = ":serial_number",
)

spdx(
    name = "app_spdx_document_namespace",
    document_namespace = ":document_namespace",
    format = "json-ld",
    sbom = ":app_sbom",
)

spdx(
    name = "app_spdx23_document_namespace",
    document_namespace = ":document_namespace",
    sbom = ":app23_sbom",
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

sbom_golden_test(
    name = "app_cyclonedx_serial_number_golden_test",
    golden = "testdata/app_cyclonedx_serial_number.json",
    report = ":app_cyclonedx_serial_number",
)

sbom_golden_test(
    name = "app_spdx_document_namespace_golden_test",
    golden = "testdata/app_spdx_document_namespace.json",
    report = ":app_spdx_document_namespace",
)

sbom_golden_test(
    name = "app_spdx23_document_namespace_golden_test",
    golden = "testdata/app_spdx23_document_namespace.json",
    report = ":app_spdx23_document_namespace",
)

strict_validations_test_suite(
    name = "strict_validations_tests",
    cyclonedx = ":app_cyclonedx",
    spdx = ":app_spdx",
)

# SPDX 2.3 has no way to express a custom exception, and the report must not
# silently weaken the license, so the exporter fails. `app_sbom` has one.
spdx_expected_failure(
    name = "custom_exception_spdx23",
    format = "json",
    message = "cannot be represented in SPDX 2.3",
    sbom = ":app_sbom",
)

build_test(
    name = "custom_exception_spdx23_test",
    targets = [":custom_exception_spdx23"],
)

# The time of a report is the Unix epoch unless the build is stamped. The golden
# tests pin it, so these look at the reports as they are generated.
sbom_timestamp_test(
    name = "app_cyclonedx_timestamp_test",
    path = ".metadata.timestamp",
    report = ":app_cyclonedx",
)

sbom_timestamp_test(
    name = "app_spdx23_timestamp_test",
    path = ".creationInfo.created",
    report = ":app_spdx23",
)

sbom_timestamp_test(
    name = "app_spdx_timestamp_test",
    path = '[.. | objects | select(has("created")) | .created] | first',
    report = ":app_spdx",
)

# A serial number or a document namespace that is not valid is an error, and
# nothing is written.
[
    [
        write_file(
            name = "{}_serial_number".format(case),
            out = "{}_serial_number.txt".format(case),
            content = [content],
        ),
        cyclonedx_expected_failure(
            name = "{}_serial_number_cyclonedx".format(case),
            format = "json",
            message = "must contain a serial number in lowercase urn:uuid form",
            sbom = ":app_sbom",
            serial_number = ":{}_serial_number".format(case),
        ),
        build_test(
            name = "{}_serial_number_test".format(case),
            targets = [":{}_serial_number_cyclonedx".format(case)],
        ),
    ]
    for case, content in [
        ("malformed", "not-a-serial-number"),
        # The CycloneDX schema only accepts lowercase hexadecimal digits.
        ("uppercase", "urn:uuid:7C9E6679-7425-40DE-944B-E07FC1F90AE7"),
    ]
]

[
    [
        write_file(
            name = "{}_document_namespace".format(case),
            out = "{}_document_namespace.txt".format(case),
            content = content,
        ),
        [
            spdx_expected_failure(
                name = "{}_document_namespace_{}".format(case, suffix),
                document_namespace = ":{}_document_namespace".format(case),
                format = format,
                message = message,
                sbom = ":app_sbom",
            )
            for suffix, format in {
                "spdx23": "json",
                "spdx301": "json-ld",
            }.items()
        ],
        build_test(
            name = "{}_document_namespace_test".format(case),
            targets = [
                ":{}_document_namespace_spdx23".format(case),
                ":{}_document_namespace_spdx301".format(case),
            ],
        ),
    ]
    for case, content, message in [
        ("empty", [], "must not be empty"),
        ("relative", ["example/spdxdocs/app"], "absolute URI"),
        ("fragment", ["https://example.com/spdxdocs/app#document"], "fragment"),
    ]
]

# The examples: a graph mixing every kind of license declaration.
[
    sbom_golden_test(
        name = "examples_{}_golden_test".format(report),
        golden = "testdata/examples_{}.json".format(report),
        report = "@supply_chain_examples//sbom:{}".format(report),
    )
    for report in [
        "app_cyclonedx",
        "app_spdx_jsonld",
        "lib_with_third_party_deps_cyclonedx",
        "lib_with_third_party_deps_cyclonedx_no_patch_files",
        "lib_with_third_party_deps_spdx",
        "lib_with_third_party_deps_spdx_jsonld",
    ]
]

strict_component_types_test_suite(
    name = "strict_component_types_tests",
    typed_sbom = "//sbom/edge_cases:service_sbom",
    untyped_sbom = "//sbom/edge_cases:image_sbom",
)
