"""Normalization for Hackage PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/hackage-definition.json
"""

visibility([
    "//purl/private/normalization/...",
])

def normalize_hackage(components):
    """Normalizes Hackage PURL components."""
    components["name"] = components["name"].replace("_", "-").replace(" ", "-")
    return components
