"""Declares provider `PackageMetadataToolchainInfo`."""

visibility("public")

def _init(metadata_overrides = []):
    return {
        "metadata_overrides": metadata_overrides,
    }

PackageMetadataToolchainInfo, _create = provider(
    doc = """
Toolchain for `package_metadata`.
""".strip(),
    fields = {
        "metadata_overrides": """
A sequence of `PackageMetadataOverrideInfo` providers.
""".strip(),
    },
    init = _init,
)
