"""Providers for transitively gathering all license and package_info targets.

Warning: This is private to the aspect that walks the tree. The API is subject
to change at any release.
"""

TargetWithMetadataInfo = provider(
    doc = """A target and the assocated metadata for it.""",
    fields = {
        "target": "Label: A target which will be associated with some metadata.",
        "metadata": "depset(): [list] of my direct collected leaf attribute providers",
        "direct_deps": "list<Label>: Direct dependencies of this target (for graph edges)",
        "direct_dep_involvements": "tuple<string>: Involvement values parallel to direct_deps",
        "direct_dep_relationships": "tuple<string>: Relationship values parallel to direct_deps",
    },
)

TransitiveMetadataInfo = provider(
    doc = """The transitive set of TargetWithMetadataInfo objects.""",
    fields = {
        "transitive": "depset(): transitive collection of TWMI",
        "top_level_target": "Label: The top level target label we are examining.",
        "is_exec_config": "bool: Whether the top-level target is in an exec configuration.",
    },
)

null_transitive_metadata_info = TransitiveMetadataInfo(
    transitive = depset(),
    top_level_target = None,
    is_exec_config = False,
)
