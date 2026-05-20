"""Declares provider `TargetMetadataInfo`."""

visibility("public")

def _init(metadata, files = []):
    return {
        "files": depset(
            direct = [
                metadata,
            ],
            transitive = files,
        ),
        "metadata": metadata,
    }

TargetMetadataInfo, _create = provider(
    doc = """
Provider for declaring metadata about a Bazel target.

> **Fields in this provider are not covered by the stability guarantee.**
""".strip(),
    fields = {
        "files": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
[File](https://bazel.build/rules/lib/builtins/File)s with metadata about the
target, including transitive files from all dependencies.
""".strip(),
        "metadata": """
The [File](https://bazel.build/rules/lib/builtins/File) containing metadata
about the target.
""".strip(),
    },
    init = _init,
)
