"""Utilities for retrieving package metadata for a target."""

visibility([
    "//common/...",
])

def get_package_metadata(*, toolchain, label, package_metadata):
    """Returns the `PackageMetadataInfo` providers that apply to a target.

    Metadata declared by a target (usually via its `package_metadata` attribute)
    can be overridden by the consumer of a module, e.g., to attach metadata to a
    dependency that does not declare any itself. Such overrides are registered as
    `PackageMetadataOverrideInfo` providers on the `package_metadata` toolchain,
    each of them declaring the set of packages it applies to.

    This function resolves both sources into the metadata to use for `label`: if
    any override on `toolchain` matches `label`, the metadata of all matching
    overrides is returned and the metadata declared by the target itself is
    ignored. Otherwise, `package_metadata` is returned unchanged.

    Example:

    ```starlark
    load("@package_metadata//common:common.bzl", "package_metadata_common")
    load("@package_metadata//providers:package_metadata_info.bzl", "PackageMetadataInfo")

    _TOOLCHAIN_TYPE = "@package_metadata//toolchains:toolchain_type"

    def _my_rule_impl(ctx):
        metadata = package_metadata_common.get_package_metadata(
            toolchain = ctx.toolchains[_TOOLCHAIN_TYPE],
            label = ctx.label,
            package_metadata = [m[PackageMetadataInfo] for m in ctx.attr.package_metadata],
        )
        # ...

    my_rule = rule(
        implementation = _my_rule_impl,
        attrs = {
            "package_metadata": attr.label_list(
                providers = [PackageMetadataInfo],
            ),
        },
        toolchains = [
            config_common.toolchain_type(_TOOLCHAIN_TYPE, mandatory = False),
        ],
    )
    ```

    Args:
        toolchain: The resolved `package_metadata`
            [toolchain](https://bazel.build/extending/toolchains), whose
            `package_metadata` field is a `PackageMetadataToolchainInfo`
            provider declaring the registered overrides. May be `None` if the
            toolchain is optional and was not resolved, in which case no
            overrides are applied.
        label: The [Label](https://bazel.build/rules/lib/builtins/Label) of the
            target to look up metadata for. It is matched against the packages
            each override applies to.
        package_metadata: A list of `PackageMetadataInfo` providers declared by
            the target itself, used if no override applies to `label`.

    Returns:
        A list of `PackageMetadataInfo` providers: the metadata of all overrides
        matching `label`, or `package_metadata` if there is no such override.
    """

    if toolchain:
        metadata = []
        for metadata_override in toolchain.package_metadata.metadata_overrides:
            if metadata_override.packages.contains(label):
                metadata.append(metadata_override.metadata)

        if metadata:
            return metadata

    return package_metadata
