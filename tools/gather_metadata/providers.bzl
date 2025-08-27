"""Providers for transitively gathering all license and package_info targets.

Warning: This is private to the aspect that walks the tree. The API is subject
to change at any release.
"""

TransitiveMetadataInfo = provider(
    doc = """The transitive set of metadata applicable to a target.""",
    fields = {
        "top_level_target": "Label: The top level target label we are examining.",
        "metadata": "depset()",

        "target": "Label: A target which will be associated with some metadata.",
        # NOT NEEDED "deps": "depset(provider): The transitive list of dependencies that ar metadata.",
        "traces": "list(string) - diagnostic for tracing a dependency relationship to a target.",
    },
)

null_transitive_metadata_info = TransitiveMetadataInfo()
