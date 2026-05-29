"""Normalization for Hackage PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/hackage-definition.json
"""

visibility([
    "//private/normalization/...",
])

def normalize_hackage(components):
    # https://github.com/package-url/purl-spec/blob/main/types/hackage-definition.json#L19-L21
    # "Apply kebab-case"
    components["name"] = components["name"].replace("_", "-").replace(" ", "-")
    return components
