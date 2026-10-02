"""Declares provider `ProvenanceInfo`."""

visibility("public")

def _init(provenance, files = []):
    return {
        "files": depset(
            direct = [
                provenance,
            ],
            transitive = files,
        ),
        "provenance": provenance,
    }

ProvenanceInfo, _create = provider(
    doc = """
Provider for describing provenance info.

This is typically emitted by every target in the dependency graph of an artifact
and contains the `PackageMetadataInfo`s directly attached to the target as well
as `ProvenanceInfo` from dependencies of the target and their relationship to
the current `ProvenanceInfo`.

`ProvenanceInfo` provides information about a single node in the
(configured) target graph, including outgoing edges to its direct dependencies.
""".strip(),
    fields = {
        "files": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
[File](https://bazel.build/rules/lib/builtins/File)s with metadata about the
target, including transitive files from all dependencies.
""".strip(),
        "provenance": """
The [File](https://bazel.build/rules/lib/builtins/File) containing the
provenance information.
""".strip(),
    },
    init = _init,
)
