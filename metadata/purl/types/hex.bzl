"""Validation for hex PURLs."""

load("//purl/private/validation:helpers.bzl", "validate_defined_type")

visibility("public")

def validate_hex(*, type, namespace, name, version, qualifiers, subpath):
    return validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)
