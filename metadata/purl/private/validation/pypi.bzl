"""Validation for PyPI PURLs.

Spec: https://github.com/package-url/purl-spec/blob/main/types/pypi-definition.json
"""

visibility([
    "//purl/private/validation/...",
])

def validate_pypi(*, type, namespace, name, version, qualifiers, subpath, strict):
    """Validates PyPI PURLs."""
    if namespace:
        return "PyPI PURLs must not have a namespace"

    return None
