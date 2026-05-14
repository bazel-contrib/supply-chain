"""Shared validation for PURL types without extra structural rules."""

visibility([
    "//purl/private/validation/...",
])

def validate_identity(*, type, namespace, name, version, qualifiers, subpath):
    return None
