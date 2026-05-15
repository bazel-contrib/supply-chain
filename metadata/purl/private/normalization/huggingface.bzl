"""Normalization for Hugging Face PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/huggingface-definition.json
"""

visibility([
    "//purl/private/normalization/...",
])

def normalize_huggingface(components):
    components["version"] = components["version"].lower() if components["version"] else components["version"]
    return components
