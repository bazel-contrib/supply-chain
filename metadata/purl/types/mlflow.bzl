"""Validation for mlflow PURLs."""

load("//purl/private/validation:helpers.bzl", "validate_defined_type")

visibility("public")

def validate_mlflow(*, type, namespace, name, version, qualifiers, subpath):
    return validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def normalize_mlflow(*, type, namespace, name, version, qualifiers, subpath):
    repository_url = qualifiers.get("repository_url") if qualifiers else None
    normalized_name = name
    if repository_url and "databricks" in repository_url.lower():
        normalized_name = name.lower()

    return struct(
        type = type,
        namespace = namespace,
        name = normalized_name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
