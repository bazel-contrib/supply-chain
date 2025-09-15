"""Declares provider `PackageMetadataToolchainInfo`."""

visibility("public")

def _init(metadata_overrides = []):
    return {
        "metadata_overrides": metadata_overrides,
    }

PackageMetadataToolchainInfo, _create = provider(
    doc = """
Toolchain for `package_metadata`.

> **Fields in this provider are not covered by the stability gurantee.**
""".strip(),
    fields = {
        "metadata_overrides": """
A sequence of `PackageMetadataOverrideInfo` providers.
""".strip(),
    },
    init = _init,
)
