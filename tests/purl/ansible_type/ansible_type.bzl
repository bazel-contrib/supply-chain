"""Ansible PURL type registration used by module extension tests."""

load("@package_metadata//purl:providers.bzl", "PurlTypeInfo")

_ANSIBLE_QUALIFIERS = {
    "download_url": True,
    "packaging": True,
    "repository_url": True,
    "vcs_url": True,
}

def _validate_ansible(*, type, namespace, name, version, qualifiers, subpath):
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

def _normalize_ansible(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type.lower(),
        namespace = namespace.lower() if namespace else namespace,
        name = name.lower() if name else name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )

def _ansible_type_impl(ctx):
    return [
        PurlTypeInfo(
            validate = _validate_ansible,
            normalize = _normalize_ansible,
        ),
    ]

ansible_type = rule(
    implementation = _ansible_type_impl,
    provides = [
        PurlTypeInfo,
    ],
)
