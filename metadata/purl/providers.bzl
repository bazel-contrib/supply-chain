"""Providers for PURL extension points."""

visibility("public")

PurlTypeInfo = provider(
    doc = "Validation and normalization functions for a PURL type.",
    fields = {
        "normalize": "Function that normalizes the PURL components for this type.",
        "validate": "Function that validates the PURL components for this type.",
    },
)
