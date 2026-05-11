"""Validation for OTP PURLs.

Spec: https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/otp-definition.md
"""

load("//purl/private/validation:helpers.bzl", "validate_with_specific")

visibility("public")

def _validate_otp_specific(*, type, namespace, name, version, qualifiers, subpath):
    """Validates OTP PURLs.

    OTP PURLs must NOT have a namespace component.

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
    # Spec requirement: "The component is unused and MUST be empty"
    # https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/otp-definition.md#L23-L24
    if namespace:
        return "OTP PURLs must not have a namespace"

    return None

def validate_otp(*, type, namespace, name, version, qualifiers, subpath):
    return validate_with_specific(type, _validate_otp_specific, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)
