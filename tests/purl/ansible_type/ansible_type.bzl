"""Ansible PURL type registration used by module extension tests."""

_ANSIBLE_QUALIFIERS = {
    "download_url": True,
    "packaging": True,
    "repository_url": True,
    "vcs_url": True,
}

def validate_ansible(*, type, namespace, name, version, qualifiers, subpath):
    if type != "ansible":
        return "Expected ansible PURL type, got {}".format(type)
    if not namespace:
        return "Ansible PURLs require a namespace"
    if not name:
        return "Mandatory property 'name' not set"

    if qualifiers:
        for key in qualifiers:
            if key not in _ANSIBLE_QUALIFIERS:
                return "Ansible qualifier '{}' is not supported".format(key)

    return None

def normalize_ansible(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type.lower(),
        namespace = namespace.lower() if namespace else namespace,
        name = name.lower() if name else name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
