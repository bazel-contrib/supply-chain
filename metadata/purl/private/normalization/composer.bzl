"""Normalization for Composer PURLs."""

visibility([
    "//purl/private/normalization/...",
])

def normalize_composer(components):
    components["namespace"] = components["namespace"].lower() if components["namespace"] else components["namespace"]
    components["name"] = components["name"].lower() if components["name"] else components["name"]
    return components
