"""Declares provider `PackageMetadataToolchainInfo`."""

visibility("public")

def _init(overrides = []):
    return {
        "overrides": {o.package: o for o in overrides},
    }

PackageMetadataToolchainInfo, _create = provider(
    doc = """

""".strip(),
    fields = {
        "overrides": """
""".strip(),
    },
    init = _init,
)
