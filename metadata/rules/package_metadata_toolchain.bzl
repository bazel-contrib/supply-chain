"""Declares rule `package_metadata_toolchain`."""

load("//providers:package_metadata_override_info.bzl", "PackageMetadataOverrideInfo")
load("//providers:package_metadata_toolchain_info.bzl", "PackageMetadataToolchainInfo")

visibility("public")

def _package_metadata_toolchain_impl(ctx):
    info = PackageMetadataToolchainInfo(
        metadata_overrides = [override[PackageMetadataOverrideInfo] for override in ctx.attr.overrides],
    )

    return [
        info,
        platform_common.ToolchainInfo(
            package_metadata = info,
        ),
    ]

# Documentation lives on the public `package_metadata_toolchain` macro below: Stardoc
# renders the macro, not this private rule.
_package_metadata_toolchain = rule(
    implementation = _package_metadata_toolchain_impl,
    attrs = {
        "overrides": attr.label_list(
            mandatory = False,
            providers = [
                PackageMetadataOverrideInfo,
            ],
        ),
    },
    provides = [
        platform_common.ToolchainInfo,
        PackageMetadataToolchainInfo,
    ],
)

def package_metadata_toolchain(
        # Disallow unnamed attributes.
        *,
        # `_package_metadata_toolchain` attributes.
        name,
        overrides = [],
        # Common attributes (subset since this target is non-configurable).
        visibility = None,
        tags = None):
    """Rule for declaring the `package_metadata` toolchain.

    The toolchain carries the consumer-side configuration of
    `package_metadata`: the `PackageMetadataOverrideInfo` providers of all
    `overrides` that the consumer of a module wants to apply to the build.

    Consumers do not read the toolchain directly. They declare it, register it
    with `register_toolchains`, and rule implementations resolve the metadata of
    a target with `package_metadata_common.get_package_metadata`, which consults
    the registered `overrides` before falling back to the metadata a target
    declares itself.

    Usage:

    ```starlark
    load("@package_metadata//rules:package_metadata_toolchain.bzl", "package_metadata_toolchain")
    load("@package_metadata_extensions//rules:package_metadata_override.bzl", "package_metadata_override")

    package_metadata_override(
        name = "legacy_dep_metadata",
        metadata = ":legacy_dep_package_metadata",
        targets = ["@legacy_dep//..."],
    )

    package_metadata_toolchain(
        name = "package_metadata_toolchain",
        overrides = [
            ":legacy_dep_metadata",
        ],
    )

    toolchain(
        name = "toolchain",
        toolchain = ":package_metadata_toolchain",
        toolchain_type = "@package_metadata//toolchains:toolchain_type",
    )
    ```

    ```starlark
    # MODULE.bazel
    register_toolchains("//:toolchain")
    ```

    The target provides `PackageMetadataToolchainInfo` as well as
    [ToolchainInfo](https://bazel.build/rules/lib/providers/ToolchainInfo) with
    the same provider in its `package_metadata` field, so it is usable both as
    the `toolchain` of a
    [toolchain](https://bazel.build/reference/be/platforms-and-toolchains#toolchain)
    target and as a plain dependency.

    Since the overrides of a build are a property of the build and not of the
    configuration, this target is non-configurable: `overrides` does not accept
    `select`.

    Args:
      name: A unique name for this target.
      overrides: A list of overrides to apply to the build.

        Each element must be a target providing
        `PackageMetadataOverrideInfo` (e.g., a `package_metadata_override`
        target), declaring the metadata to use and the packages it applies to.

        An empty list declares a toolchain that applies no overrides, i.e., all
        targets keep the metadata they declare themselves.
      visibility: The visibility of this target.

        Registering a toolchain does not require it to be visible to the
        packages it affects, so this is typically left at its default.
      tags: A list of arbitrary tags to apply to this target.
    """

    _package_metadata_toolchain(
        # `_package_metadata_toolchain` attributes.
        name = name,
        overrides = overrides,

        # Common attributes.
        visibility = visibility,
        tags = tags,
        applicable_licenses = [],
    )
