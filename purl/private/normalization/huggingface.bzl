"""Normalization for Hugging Face PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/huggingface-definition.json
"""

visibility([
    "//private/normalization/...",
])

def normalize_huggingface(components):
    # While these are not described in the normalization section of the spec, they are validated as such in the
    # type specific tests
    components["version"] = components["version"].lower() if components["version"] else components["version"]
    return components
