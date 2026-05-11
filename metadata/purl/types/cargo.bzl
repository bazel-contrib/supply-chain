"""Validation for cargo PURLs."""

load("//purl/types:helpers.bzl", "validate_defined_type")

visibility("public")

def validate_cargo(*, type, namespace, name, version, qualifiers, subpath):
    return validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)
