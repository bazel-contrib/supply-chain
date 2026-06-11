"""Normalization for PyPI PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/pypi-definition.json
"""

visibility([
    "//purl/private/normalization/...",
])

def normalize_pypi(components):
    # https://github.com/package-url/purl-spec/blob/main/types/pypi-definition.json#L20-L23
    # "Replace underscore _ with dash -",
    # "Replace dot . with underscore _ when used in distribution (sdist, wheel) names"
    components["name"] = components["name"].lower().replace("_", "-").replace(".", "_")
    return components
