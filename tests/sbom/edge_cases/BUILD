# SBOM reports of graphs with unusual shapes, checked against goldens.
# See ../golden_test.bzl to update them.

load("@package_metadata//attributes:component_type.bzl", "component_type")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@supply_chain_tools//sbom:cyclonedx.bzl", "cyclonedx")
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

# Packages declare their component type; one that does not keep the default.
[
    [
        component_type(
            name = "{}_type".format(name),
            type = type,
        ),
        package_metadata(
            name = "{}_metadata".format(name),
            attributes = [":{}_type".format(name)],
            purl = purl,
        ),
    ]
    for name, type, purl in [
        ("image", "container", "pkg:oci/my-image?repository_url=registry.example.com%2Fmy-image"),
        ("base_os", "operating-system", "pkg:generic/base-os@12.1"),
        ("bootloader", "firmware", "pkg:generic/bootloader@2.4"),
        ("service", "application", "pkg:generic/service@3.0.0"),
    ]
]

[
    filegroup(
        name = name,
        applicable_licenses = [":{}_metadata".format(name)],
    )
    for name in [
        "base_os",
        "bootloader",
        "service",
    ]
]

# Without a component type, so it is a library.
filegroup(
    name = "image_library",
    applicable_licenses = [":direct_metadata"],
)

filegroup(
    name = "image",
    srcs = [
        ":base_os",
        ":bootloader",
        ":image_library",
        ":service",
    ],
    applicable_licenses = [":image_metadata"],
)

sbom(
    name = "image_sbom",
    target = ":image",
)

[
    [
        spdx(
            name = "image_{}".format(name),
            format = format,
            sbom = ":image_sbom",
        ),
        sbom_golden_test(
            name = "image_{}_golden_test".format(name),
            golden = "testdata/image_{}.json".format(name),
            report = ":image_{}".format(name),
        ),
    ]
    for name, format in {
        "spdx": "json-ld",
        "spdx_2_3": "json",
    }.items()
]

cyclonedx(
    name = "image_cyclonedx",
    sbom = ":image_sbom",
)

sbom_golden_test(
    name = "image_cyclonedx_golden_test",
    golden = "testdata/image_cyclonedx.json",
    report = ":image_cyclonedx",
)
