"""Providers for transitively gathering all license and package_info targets.

Warning: This is private to the aspect that walks the tree. The API is subject
to change at any release.
"""

TransitiveMetadataInfo = provider(
    doc = """The transitive set of metadata applicable to a target.""",
    fields = {
        "target": "Label: A target which will be associated with some metadata.",
	# only one of directs or trans may be filled in
        "directs": "depset(): [list] of my direct collected leaf providers",
        "trans": "depset(): transitive collection of TMI",
        "top_level_target": "Label: The top level target label we are examining.",
        "traces": "list(string) - diagnostic for tracing a dependency relationship to a target.",
    },
)

null_transitive_metadata_info = TransitiveMetadataInfo()
