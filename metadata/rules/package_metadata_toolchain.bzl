"""Declares rule `package_metadata_toolchain`."""

load("//providers:package_attribute_override_info.bzl", "PackageAttributeOverrideInfo")
load("//providers:package_metadata_toolchain_info.bzl", "PackageMetadataToolchainInfo")

visibility("public")

def _package_metadata_toolchain_impl(ctx):
    return [
        platform_common.ToolchainInfo(
            package_metadata = PackageMetadataToolchainInfo(
                overrides = [o[PackageAttributeOverrideInfo] for o in ctx.attr.overrides],
            ),
        ),
    ]

_package_metadata_toolchain = rule(
    implementation = _package_metadata_toolchain_impl,
    attrs = {
        "overrides": attr.label_list(
            mandatory = False,
            doc = """
""".strip(),
            providers = [
                PackageAttributeOverrideInfo,
            ],
        ),
    },
    provides = [
        platform_common.ToolchainInfo,
    ],
    doc = """
Rule for declaring overrides for a `package_metadata` attribute.
""".strip(),
)

def package_metadata_toolchain(
        # Disallow unnamed attributes.
        *,
        # `_package_metadata_toolchain` attributes.
        name,
        overrides = [],
        # Common attributes (subset since this target is non-configurable).
        visibility = None):
    _package_metadata_toolchain(
        # `_package_metadata_toolchain` attributes.
        name = name,
        overrides = overrides,

        # Common attributes.
        visibility = visibility,
        applicable_licenses = [],
    )
