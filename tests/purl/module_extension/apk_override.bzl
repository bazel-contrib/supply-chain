"""APK PURL type override used by module extension tests."""

_APK_QUALIFIERS = {
    "arch": True,
    "distro": True,
    "upstream": True,
}

def validate_apk(*, type, namespace, name, version, qualifiers, subpath):
    if type != "apk":
        return "Expected apk PURL type, got {}".format(type)
    if not namespace:
        return "APK PURLs require a namespace"
    if not name:
        return "Mandatory property 'name' not set"

    if qualifiers:
        for key in qualifiers:
            if key not in _APK_QUALIFIERS:
                return "APK qualifier '{}' is not supported".format(key)

    return None

def normalize_apk(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type.lower(),
        namespace = namespace.lower() if namespace else namespace,
        name = name.lower() if name else name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
