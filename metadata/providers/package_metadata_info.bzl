"""Declares provider `PackageMetadataInfo`."""

visibility("public")

def _init(metadata, files = [], attributes = []):
    return {
        "attributes": attributes,
        "files": depset(
            direct = [
                metadata,
            ],
            transitive = files,
        ),
        "metadata": metadata,
    }

PackageMetadataInfo, _create = provider(
    doc = """
Provider for declaring metadata about a Bazel package.

> **Fields in this provider are not covered by the stability gurantee.**
""".strip(),
    fields = {
        "attributes": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
`PackageAttributeInfo` providers.
""".strip(),
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
    init = _init,
)
