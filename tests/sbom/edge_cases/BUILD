# Reports of graphs with unusual shapes, checked against goldens.
# See ../golden_test.bzl to update them.

load("@bazel_skylib//rules:build_test.bzl", "build_test")
load("@package_metadata//attributes:cpe.bzl", "cpe")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@supply_chain_tools//sbom:cyclonedx.bzl", "cyclonedx")
load("@supply_chain_tools//sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//sbom:spdx.bzl", "spdx")
load("//sbom:expected_failure.bzl", "spdx_expected_failure")
load("//sbom:golden_test.bzl", "sbom_golden_test")

package_metadata(
    name = "app_metadata",
    purl = "pkg:generic/app@1.0.0",
)

# A package url without a version, and with each of the download qualifiers.
package_metadata(
    name = "unversioned_metadata",
    purl = "pkg:generic/unversioned",
)

package_metadata(
    name = "download_url_metadata",
    purl = "pkg:generic/download_url@1.2.3?download_url=https://example.com/download_url.tar.gz",
)

package_metadata(
    name = "url_download_metadata",
    purl = "pkg:generic/url_download@1.2.3?url_download=https://example.com/url_download.tar.gz",
)

filegroup(
    name = "unversioned",
    applicable_licenses = [":unversioned_metadata"],
)

filegroup(
    name = "download_url",
    applicable_licenses = [":download_url_metadata"],
)

filegroup(
    name = "url_download",
    applicable_licenses = [":url_download_metadata"],
)

filegroup(
    name = "qualifiers",
    srcs = [
        ":download_url",
        ":unversioned",
        ":url_download",
    ],
    applicable_licenses = [":app_metadata"],
)

sbom(
    name = "qualifiers_sbom",
    target = ":qualifiers",
)

spdx(
    name = "qualifiers_spdx",
    format = "json-ld",
    sbom = ":qualifiers_sbom",
)

sbom_golden_test(
    name = "qualifiers_spdx_golden_test",
    golden = "testdata/qualifiers_spdx.json",
    report = ":qualifiers_spdx",
)

# Targets without metadata between packages (wrappers) and a package reached
# by several paths: its dependency edges must be kept, once each.
package_metadata(
    name = "direct_metadata",
    purl = "pkg:generic/direct@1.0.0",
)

package_metadata(
    name = "indirect_metadata",
    purl = "pkg:generic/indirect@1.0.0",
)

filegroup(
    name = "indirect",
    applicable_licenses = [":indirect_metadata"],
)

filegroup(
    name = "wrapper_a",
    srcs = [":indirect"],
)

filegroup(
    name = "wrapper_b",
    srcs = [":indirect"],
)

filegroup(
    name = "direct",
    srcs = [
        ":wrapper_a",
        ":wrapper_b",
    ],
    applicable_licenses = [":direct_metadata"],
)

filegroup(
    name = "root_wrapper",
    srcs = [":direct"],
)

filegroup(
    name = "wrapped",
    srcs = [":root_wrapper"],
    applicable_licenses = [":app_metadata"],
)

sbom(
    name = "wrapped_sbom",
    target = ":wrapped",
)

spdx(
    name = "wrapped_spdx",
    format = "json-ld",
    sbom = ":wrapped_sbom",
)

sbom_golden_test(
    name = "wrapped_spdx_golden_test",
    golden = "testdata/wrapped_spdx.json",
    report = ":wrapped_spdx",
)

# No metadata on the root: the document is named after the first package.
filegroup(
    name = "no_root",
    srcs = [":direct"],
)

sbom(
    name = "no_root_sbom",
    require_root_metadata = False,
    target = ":no_root",
)

spdx(
    name = "no_root_spdx",
    format = "json-ld",
    sbom = ":no_root_sbom",
)

sbom_golden_test(
    name = "no_root_spdx_golden_test",
    golden = "testdata/no_root_spdx.json",
    report = ":no_root_spdx",
)

# No metadata at all.
filegroup(
    name = "empty",
)

sbom(
    name = "empty_sbom",
    require_root_metadata = False,
    target = ":empty",
)

spdx(
    name = "empty_spdx",
    format = "json-ld",
    sbom = ":empty_sbom",
)

sbom_golden_test(
    name = "empty_spdx_golden_test",
    golden = "testdata/empty_spdx.json",
    report = ":empty_spdx",
)

# A dependency on a generated file: the target lists the output of a rule with
# package metadata (like a java_binary and its _deploy.jar), here two of them.
# The package must be reported as a dependency of the target, once, and not left
# out of its dependency graph.
package_metadata(
    name = "generated_metadata",
    purl = "pkg:generic/generated@1.0.0",
)

genrule(
    name = "generated",
    outs = [
        "generated.txt",
        "generated_other.txt",
    ],
    applicable_licenses = [":generated_metadata"],
    cmd = "echo generated > $(location generated.txt) && echo other > $(location generated_other.txt)",
)

filegroup(
    name = "generated_file_app",
    srcs = [
        ":generated.txt",
        ":generated_other.txt",
    ],
    applicable_licenses = [":app_metadata"],
)

sbom(
    name = "generated_file_sbom",
    target = ":generated_file_app",
)

spdx(
    name = "generated_file_spdx",
    format = "json-ld",
    sbom = ":generated_file_sbom",
)

sbom_golden_test(
    name = "generated_file_spdx_golden_test",
    golden = "testdata/generated_file_spdx.json",
    report = ":generated_file_spdx",
)

# A root without dependencies, alone in the report: CycloneDX has to declare it
# with an empty dependency list, since a component absent from the dependency
# graph has unknown dependencies.
filegroup(
    name = "root_only",
    applicable_licenses = [":app_metadata"],
)

sbom(
    name = "root_only_sbom",
    target = ":root_only",
)

cyclonedx(
    name = "root_only_cyclonedx",
    sbom = ":root_only_sbom",
)

sbom_golden_test(
    name = "root_only_cyclonedx_golden_test",
    golden = "testdata/root_only_cyclonedx.json",
    report = ":root_only_cyclonedx",
)

# The CPEs of a package. SPDX reports every one of them, as an external
# reference (2.3) or an external identifier (3.0.1) of the type of its CPE
# version. CycloneDX has room for a single CPE, so it reports the first one
# and keeps the others as properties.
cpe(
    name = "cpe_single",
    identifiers = ["cpe:2.3:a:acme:single:1.0.0:*:*:*:*:*:*:*"],
)

cpe(
    name = "cpe_many",
    identifiers = [
        "cpe:2.3:a:acme:many:1.0.0:*:*:*:*:*:*:*",
        "cpe:/a:acme:many:1.0.0",
        "cpe:2.3:a:acme_inc:many:1.0.0:*:*:*:*:*:*:*",
    ],
)

package_metadata(
    name = "single_cpe_metadata",
    attributes = [":cpe_single"],
    purl = "pkg:generic/single@1.0.0",
)

package_metadata(
    name = "many_cpes_metadata",
    attributes = [":cpe_many"],
    purl = "pkg:generic/many@1.0.0",
)

filegroup(
    name = "single_cpe",
    applicable_licenses = [":single_cpe_metadata"],
)

filegroup(
    name = "many_cpes",
    applicable_licenses = [":many_cpes_metadata"],
)

# A package without a CPE, next to the ones with.
filegroup(
    name = "cpe_app",
    srcs = [
        ":download_url",
        ":many_cpes",
        ":single_cpe",
    ],
    applicable_licenses = [":app_metadata"],
)

sbom(
    name = "cpe_sbom",
    target = ":cpe_app",
)

[
    [
        spdx(
            name = "cpe_{}".format(name),
            format = format,
            sbom = ":cpe_sbom",
        ),
        sbom_golden_test(
            name = "cpe_{}_golden_test".format(name),
            golden = "testdata/cpe_{}.json".format(name),
            report = ":cpe_{}".format(name),
        ),
    ]
    for name, format in {
        "spdx": "json-ld",
        "spdx_2_3": "json",
    }.items()
]

cyclonedx(
    name = "cpe_cyclonedx",
    sbom = ":cpe_sbom",
)

sbom_golden_test(
    name = "cpe_cyclonedx_golden_test",
    golden = "testdata/cpe_cyclonedx.json",
    report = ":cpe_cyclonedx",
)

# A CPE that is neither a CPE 2.2 URI nor a CPE 2.3 formatted string is an error.
cpe(
    name = "cpe_malformed",
    identifiers = ["cpe:2.3:a:acme"],
)

package_metadata(
    name = "malformed_cpe_metadata",
    attributes = [":cpe_malformed"],
    purl = "pkg:generic/malformed@1.0.0",
)

filegroup(
    name = "malformed_cpe",
    applicable_licenses = [":malformed_cpe_metadata"],
)

sbom(
    name = "malformed_cpe_sbom",
    target = ":malformed_cpe",
)

[
    [
        spdx_expected_failure(
            name = "malformed_cpe_{}".format(name),
            format = format,
            message = "is not a CPE 2.2 URI",
            sbom = ":malformed_cpe_sbom",
        ),
        build_test(
            name = "malformed_cpe_{}_test".format(name),
            targets = [":malformed_cpe_{}".format(name)],
        ),
    ]
    for name, format in {
        "spdx": "json-ld",
        "spdx_2_3": "json",
    }.items()
]

# The organization that originated a package is inferred from the namespace of
# its purl for the hosts where the namespace is the owner of the repository
# (github, gitlab, bitbucket), and for no other type: the namespace of a maven
# group or of an npm scope is not necessarily an organization.
package_metadata(
    name = "github_metadata",
    purl = "pkg:github/acmecorp/inferred@1.0.0",
)

package_metadata(
    name = "maven_metadata",
    purl = "pkg:maven/org.acme/not_inferred@1.0.0",
)

filegroup(
    name = "github",
    applicable_licenses = [":github_metadata"],
)

filegroup(
    name = "maven",
    applicable_licenses = [":maven_metadata"],
)

filegroup(
    name = "originated",
    srcs = [
        ":github",
        ":maven",
    ],
    applicable_licenses = [":app_metadata"],
)

sbom(
    name = "originated_sbom",
    target = ":originated",
)

spdx(
    name = "originated_spdx",
    format = "json-ld",
    sbom = ":originated_sbom",
)

sbom_golden_test(
    name = "originated_spdx_golden_test",
    golden = "testdata/originated_spdx.json",
    report = ":originated_spdx",
)
