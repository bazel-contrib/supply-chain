"""Declares provider `PackageMetadataInfo`."""

PackageMetadataInfo = provider(
    doc = """
Provider for declaring metadata about a Bazel package.
""".strip(),
    fields = {
        "files": """
""".strip(),
        "metadata": """
""".strip(),
    },
)
