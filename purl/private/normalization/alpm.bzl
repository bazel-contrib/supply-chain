"""Normalization for ALPM PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/alpm-definition.json
"""

visibility([
    "//private/normalization/...",
])

def normalize_alpm(components):
    # https://github.com/package-url/purl-spec/blob/main/types/alpm-definition.json#L16-L18
    # "It is not case sensitive and must be lowercased."
    components["namespace"] = components["namespace"].lower() if components["namespace"] else components["namespace"]
    return components
