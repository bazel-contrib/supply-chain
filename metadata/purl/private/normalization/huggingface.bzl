"""Normalization for Hugging Face PURLs."""

visibility([
    "//purl/private/normalization/...",
])

def normalize_huggingface(components):
    components["version"] = components["version"].lower() if components["version"] else components["version"]
    return components
