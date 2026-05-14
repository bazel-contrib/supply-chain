"""Validation for Pub PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/pub-definition.json
"""

visibility([
    "//purl/private/validation/...",
])

def validate_pub(*, type, namespace, name, version, qualifiers, subpath):
    """Validates Pub PURLs."""
    if namespace:
        return "Pub PURLs must not have a namespace"

    return None
