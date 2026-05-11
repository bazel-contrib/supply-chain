"""Root-module APKO PURL type override."""

def validate_root_apko(*, type, namespace, name, version, qualifiers, subpath):
    if type != "apko":
        return "Expected apko PURL type, got {}".format(type)
    if not name:
        return "Mandatory property 'name' not set"
    return None

def normalize_root_apko(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type.lower(),
        namespace = namespace.lower() if namespace else namespace,
        name = name.lower() if name else name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
