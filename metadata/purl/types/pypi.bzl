"""Validation for pypi PURLs."""

load("//purl/types:helpers.bzl", "validate_defined_type")

visibility("public")

def validate_pypi(*, type, namespace, name, version, qualifiers, subpath):
    return validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _normalize_pypi_name(name):
    normalized = []
    previous_separator = False
    for c in name.elems():
        is_separator = c == "." or c == "_" or c == "-"
        if is_separator:
            if not previous_separator:
                normalized.append("-")
            previous_separator = True
        else:
            normalized.append(c.lower())
            previous_separator = False
    return "".join(normalized)

def normalize_pypi(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type,
        namespace = namespace,
        name = _normalize_pypi_name(name),
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
