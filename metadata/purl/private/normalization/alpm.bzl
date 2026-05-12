"""Normalization for ALPM PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/alpm-definition.json
"""

visibility([
    "//purl/private/normalization/...",
])

def normalize_alpm(components):
    """Normalizes ALPM PURL components."""
    components["namespace"] = components["namespace"].lower() if components["namespace"] else components["namespace"]
    return components
