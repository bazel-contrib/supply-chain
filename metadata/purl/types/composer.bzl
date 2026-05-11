"""Validation for composer PURLs."""

load("//purl/types:helpers.bzl", "validate_defined_type")

visibility("public")

def validate_composer(*, type, namespace, name, version, qualifiers, subpath):
    return validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def normalize_composer(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type,
        namespace = [segment.lower() for segment in namespace] if namespace else namespace,
        name = name.lower(),
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
