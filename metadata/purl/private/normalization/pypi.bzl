"""Normalization for PyPI PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/pypi-definition.json
"""

visibility([
    "//purl/private/normalization/...",
])

def normalize_pypi(components):
    """Normalizes PyPI PURL components."""
    components["name"] = components["name"].lower().replace("_", "-")
    return components
