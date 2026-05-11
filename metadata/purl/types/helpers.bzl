"""Generic helpers for PURL type validation."""

load("//purl/types:requirements.bzl", "TYPE_REQUIREMENTS")

visibility("public")

def _has_value(value):
    return value != None and value != "" and value != []

def _validate_type_definition(type, definition, namespace, name, version, qualifiers, subpath):
    namespace_requirement = definition.get("namespace")
    if namespace_requirement == "required" and not _has_value(namespace):
        return "{} PURLs require a namespace".format(type)
    if namespace_requirement == "prohibited" and _has_value(namespace):
        return "{} PURLs must not have a namespace".format(type)

    if not _has_value(name):
        return "Mandatory property 'name' not set"

    for key in definition.get("required_qualifiers", []):
        if not qualifiers or not _has_value(qualifiers.get(key)):
            return "{} PURLs require a '{}' qualifier".format(type, key)

    return None

def validate_defined_type(type, *, namespace, name, version, qualifiers, subpath):
    return _validate_type_definition(
        type,
        TYPE_REQUIREMENTS[type],
        namespace,
        name,
        version,
        qualifiers,
        subpath,
    )

def validate_with_specific(type, specific_validator, *, namespace, name, version, qualifiers, subpath):
    err = validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)
    if err:
        return err
    return specific_validator(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )

def identity_normalization(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
