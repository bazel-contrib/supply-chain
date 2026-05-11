"""Validation for Swift PURLs.

Spec: https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/swift-definition.md
"""

load("//purl/private/validation:helpers.bzl", "validate_with_specific")

visibility("public")

def _validate_swift_specific(*, type, namespace, name, version, qualifiers, subpath):
    """Validates Swift PURLs.

    Swift PURLs must have a namespace.

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
    # Spec requirement: Namespace is "Required" - composed of "source host and user/organization"
    # https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/swift-definition.md#L22-L24
    if not namespace:
        return "Swift PURLs require a namespace"

    return None

def validate_swift(*, type, namespace, name, version, qualifiers, subpath):
    return validate_with_specific(type, _validate_swift_specific, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)
