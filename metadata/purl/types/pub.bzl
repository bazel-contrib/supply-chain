"""Validation for pub PURLs."""

load("//purl/private/strings:strings.bzl", "strings")
load("//purl/types:helpers.bzl", "validate_with_specific")

visibility("public")

def _validate_pub_specific(*, type, namespace, name, version, qualifiers, subpath):
    # https://github.com/package-url/purl-spec/blob/8ae5b73ea3ea882aa71159703d8afab4b615622c/types/pub-definition.json#L13
    if namespace:
        return "pub PURLs must not have a namespace"

    # https://github.com/package-url/purl-spec/blob/8ae5b73ea3ea882aa71159703d8afab4b615622c/types/pub-definition.json#L19
    for c in strings.bytes.from_string(name.lower()):
        if strings.ascii.is_alphanumeric(c) or c == 95:
            continue
        return "pub names may only contain lowercase ASCII letters, numbers, and underscores"

    return None

def validate_pub(*, type, namespace, name, version, qualifiers, subpath):
    return validate_with_specific(type, _validate_pub_specific, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)
