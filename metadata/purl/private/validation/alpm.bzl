"""Validation for ALPM PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/alpm-definition.json
"""

visibility([
    "//purl/private/validation/...",
])

def validate_alpm(*, type, namespace, name, version, qualifiers, subpath):
    """Validates ALPM PURLs."""
    if not namespace:
        return "ALPM PURLs require a namespace (vendor)"

    return None
