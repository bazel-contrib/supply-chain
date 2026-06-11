"""Providers for the sbom rules.

Warning: This is private to the aspect that walks the tree. The API is subject
to change at any release.
"""

SbomInfo = provider(
    doc = "A provider that contains the graph and classifications for generating an SBOM.",
    fields = {
        "graph": "File: The graph-only JSON from gather_metadata",
        "classifications": "File: The SBOM classifications JSON from cmd/sbom",
    }
)