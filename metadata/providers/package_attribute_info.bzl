"""Declares provider `PackageAttributeInfo`."""

PackageAttributeInfo = provider(
    doc = """
Provider for declaring metadata about a Bazel package.
""".strip(),
    fields = {
        "files": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
[File](https://bazel.build/rules/lib/builtins/File)s containing information
about this attribute.
""".strip(),
        "kind": """
The identifier of the attribute.

This should generally be in reverse DNS format (e.g., `com.example.foo`).
""".strip(),
        "metadata": """
The [File](https://bazel.build/rules/lib/builtins/File) containing the metadata
of the attribute.

The format of this file depends on the `kind` of attribute. Please consult the
documentation of the attribute.
""".strip(),
    },
)
