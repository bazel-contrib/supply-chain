"""Declares provider `PackageAttributeInfo`."""

PackageAttributeInfo = provider(
    doc = """
Provider for declaring metadata about a Bazel package.
""".strip(),
    fields = {
        "files": """
""".strip(),
        "kind": """
""".strip(),
        "metadata": """
""".strip(),
    },
)
