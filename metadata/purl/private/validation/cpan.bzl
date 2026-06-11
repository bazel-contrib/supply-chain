"""Validation for CPAN PURLs.

Spec: https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/cpan-definition.md
"""

visibility([
    "//purl/private/validation/...",
])

def validate_cpan(*, type, namespace, name, version, qualifiers, subpath):
    """Validates CPAN PURLs.

    CPAN PURLs must have a namespace (author) and the name must be a distribution
    name (not a module name with ::).

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
    # Spec requirement: "It MUST be written uppercase and is required"
    # https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/cpan-definition.md#L22-L23
    if not namespace:
        return "CPAN PURLs require a namespace (author)"

    # Spec requirement: namespace "MUST be written uppercase"
    # https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/cpan-definition.md#L22
    if namespace != namespace.upper():
        return ("CPAN PURL namespace (author) must be " +
                "uppercase")

    # Spec requirement: "A distribution name MUST NOT contain the string '::'"
    # https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/cpan-definition.md#L40-L41
    if "::" in name:
        return ("CPAN PURL name must be a distribution name, " +
                "not a module name (contains '::')")

    return None
