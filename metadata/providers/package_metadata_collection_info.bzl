"""Declares provider `PackageMetadataCollectionInfo`."""

visibility("public")

def _init(dependencies = [], tools = [], deps = []):
    return {
        "dependencies": depset(
            direct = dependencies,
            transitive = [p.dependencies for p in deps],
        ),
        "files": depset(
            transitive = [p.files for p in dependencies] + [p.files for p in tools] + [p.files for p in deps],
        ),
        "tools": depset(
            direct = tools,
            transitive = [p.tools for p in deps],
        ),
    }

PackageMetadataCollectionInfo, _create = provider(
    doc = """
Provides a collection of `PackageMetadataInfo`s transitively used by target.

> **Fields in this provider are not covered by the stability gurantee.**
""".strip(),
    fields = {
        "dependencies": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
`PackageMetadataInfo`s that are dependencies of the targets.

This only includes dependencies that are used at runtime (i.e., from
`cfg = "target"`). For dependencies that are needed at compile time (e.g.,
compilers, linters, ...), see `tools`.
""".strip(),
        "files": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
[File](https://bazel.build/rules/lib/builtins/File)s referenced by
`PackageMetadataInfo` providers in this collection.
""".strip(),
        "tools": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
`PackageMetadataInfo`s that are are needed to build the target.

For dependencies that are needed at runtime time see `dependencies`.
""".strip(),
    },
    init = _init,
)
