"""Normalization for Pub PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/pub-definition.json
"""

load("//purl/private/strings:strings.bzl", "strings")

visibility([
    "//purl/private/normalization/...",
])

def normalize_pub(components):
    """Normalizes Pub PURL components."""
    result = []
    for c in strings.bytes.from_string(components["name"].lower()):
        if (c >= 97 and c <= 122) or (c >= 48 and c <= 57) or c == 95:
            result.append(c)
        else:
            result.append(95)  # _
    components["name"] = strings.bytes.to_string(result)
    return components
