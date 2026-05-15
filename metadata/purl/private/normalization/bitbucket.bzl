"""Normalization for Bitbucket PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/bitbucket-definition.json
"""

visibility([
    "//purl/private/normalization/...",
])

def normalize_bitbucket(components):
    # While these are not described in the normalization section of the spec, they are validated as such in the
    # type specific tests
    components["namespace"] = components["namespace"].lower() if components["namespace"] else components["namespace"]
    components["name"] = components["name"].lower() if components["name"] else components["name"]
    return components
