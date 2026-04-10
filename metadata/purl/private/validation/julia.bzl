"""Validation for Julia PURLs.

Spec: https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/julia-definition.md
"""

visibility([
    "//purl/private/validation/...",
])

def validate_julia(*, type, namespace, name, version, qualifiers, subpath):
    """Validates Julia PURLs.

    Julia PURLs must have either a version or a uuid qualifier.

    Args:
        type: The PURL type
        namespace: The PURL namespace
        name: The PURL name
        version: The PURL version
        qualifiers: The PURL qualifiers
        subpath: The PURL subpath

    Returns:
        An error string if validation fails, None otherwise.
    """
    has_uuid = qualifiers and qualifiers.get("uuid")

    # Spec requirement: "The Julia package UUID" is mandatory in qualifiers (OR version is provided)
    # https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/julia-definition.md#L30-L31
    # Version is optional per https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/julia-definition.md#L25-L26
    if not version and not has_uuid:
        return "Julia PURLs require either a version or a 'uuid' qualifier"

    return None
