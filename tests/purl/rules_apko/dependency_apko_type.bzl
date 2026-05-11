"""Dependency-provided APKO PURL type registration."""

_DEPENDENCY_APKO_REQUIREMENTS = {
    "namespace": "required",
    "lower_namespace": True,
    "lower_name": True,
}

def validate_dependency_apko(*, type, namespace, name, version, qualifiers, subpath):
    return "dependency apko validator"

def normalize_dependency_apko(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = "dependency-apko",
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
