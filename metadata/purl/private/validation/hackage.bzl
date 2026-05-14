"""Validation for Hackage PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/hackage-definition.json
"""

visibility([
    "//purl/private/validation/...",
])

def validate_hackage(*, type, namespace, name, version, qualifiers, subpath):
    """Validates Hackage PURLs."""
    if namespace:
        return "Hackage PURLs must not have a namespace"

    return None
