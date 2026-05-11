"""Validation for VS Code Extension PURLs.

Spec: https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/vscode-extension-definition.md
"""

load("//purl/private/validation:helpers.bzl", "validate_with_specific")

visibility("public")

def _validate_vscode_extension_specific(*, type, namespace, name, version, qualifiers, subpath):
    """Validates VS Code Extension PURLs.

    VS Code Extension PURLs must have a namespace (publisher).

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
    # Spec requirement: Namespace is "Required" - represents the publisher/vendor
    # https://github.com/package-url/purl-spec/blob/c756cacf766d4bf2711b248b935b3b80d1b1ba2e/types-doc/vscode-extension-definition.md#L21-L22
    if not namespace:
        return "VS Code Extension PURLs require a namespace (publisher)"

    return None

def validate_vscode_extension(*, type, namespace, name, version, qualifiers, subpath):
    return validate_with_specific(type, _validate_vscode_extension_specific, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)
