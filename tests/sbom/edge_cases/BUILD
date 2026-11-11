# SPDX 3.0.1 reports of graphs with unusual shapes, checked against goldens.
# See ../golden_test.bzl to update them.

load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@supply_chain_tools//sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//sbom:spdx.bzl", "spdx")
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
