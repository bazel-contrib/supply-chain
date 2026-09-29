"""Declares provider `TargetInfo`."""

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

TargetInfo, _create = provider(
    doc = """
Provider for describing a single target.

This includes the `PackageMetadataInfo`s directly attached to the target as well
as `TargetInfo` from dependencies of the target.

`TargetInfo` provides information about a single node in the
(configured) target graph, including outgoing edges to its direct dependencies.
""".strip(),
    fields = {
        "files": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
[File](https://bazel.build/rules/lib/builtins/File)s with metadata about the
target, including transitive files from all dependencies.
""".strip(),
        "info": """
The [File](https://bazel.build/rules/lib/builtins/File) containing the
information about the target.
""".strip(),
    },
    init = _init,
)
