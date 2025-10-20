"""Declares provider `PackageMetadataOverrideInfo`."""

visibility("public")

def _init(*, packages, metadata):
    return {
        "metadata": metadata,
        "packages": packages,
    }

PackageMetadataOverrideInfo, _create = provider(
    doc = """
Defines an override for `PackageMetadataInfo` for a set of packages.

> **Fields in this provider are not covered by the stability guarantee.**
""".strip(),
    fields = {
        "metadata": """
The `PackageMetadataInfo` provider to use instead of the provider declared by
package itself.
""".strip(),
        "packages": """
A [PackageSpecificationInfo](https://bazel.build/rules/lib/providers/PackageSpecificationInfo)
provider declaring which packages the override applies to.

This is typically created by a
[package_group](https://bazel.build/rules/lib/globals/build#package_group)
target.
""".strip(),
    },
    init = _init,
)
