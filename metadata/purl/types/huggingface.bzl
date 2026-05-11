"""Validation for huggingface PURLs."""

load("//purl/types:helpers.bzl", "validate_defined_type")

visibility("public")

def validate_huggingface(*, type, namespace, name, version, qualifiers, subpath):
    return validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def normalize_huggingface(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type,
        namespace = namespace,
        name = name,
        version = version.lower() if version else version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
