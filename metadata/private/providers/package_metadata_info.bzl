"""Declares provider `PackageMetadataInfo`."""

PackageMetadataInfo = provider(
    doc = """
Provider for declaring metadata about a Bazel package.
""".strip(),
    fields = {
        "files": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
[File](https://bazel.build/rules/lib/builtins/File)s with metadata about the
package, including transitive files from all attributes of the package.
""".strip(),
        "metadata": """
The [File](https://bazel.build/rules/lib/builtins/File) containing metadata
about the package.
""".strip(),
    },
)
