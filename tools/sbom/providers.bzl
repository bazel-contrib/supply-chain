"""Providers for the sbom rules.

Warning: This is private to the aspect that walks the tree. The API is subject
to change at any release.
"""

SbomInfo = provider(
    doc = "A provider that contains the configuration for generating an SBOM.",
    fields = {
        "config": "The configuration file for generating the SBOM.",
    }
)