"""Declares rule `package_metadata_toolchain`."""

load("//providers:package_metadata_override_info.bzl", "PackageMetadataOverrideInfo")
load("//providers:package_metadata_toolchain_info.bzl", "PackageMetadataToolchainInfo")

visibility("public")

def _package_metadata_toolchain_impl(ctx):
    info = PackageMetadataToolchainInfo(
        metadata_overrides = [o[PackageMetadataOverrideInfo] for o in ctx.attr.metadata_overrides],
    )

    return [
        info,
        platform_common.ToolchainInfo(
            package_metadata = info,
        ),
    ]

_package_metadata_toolchain = rule(
    implementation = _package_metadata_toolchain_impl,
    attrs = {
        "metadata_overrides": attr.label_list(
            mandatory = False,
            providers = [
                PackageMetadataOverrideInfo,
            ],
            doc = """
""".strip(),
        ),
    },
    provides = [
        PackageMetadataToolchainInfo,
        platform_common.ToolchainInfo,
    ],
    doc = """
Rule for declaring `PackageMetadataToolchainInfo`.
""".strip(),
)

def package_metadata_toolchain(
        # Disallow unnamed attributes.
        *,
        # `_package_metadata_toolchain` attributes.
        name,
        # Common attributes (subset since this target is non-configurable).
        visibility = None):
    _package_metadata_toolchain(
        # `_package_metadata_toolchain` attributes.
        name = name,

        # Common attributes.
        visibility = visibility,
        applicable_licenses = [],
    )
