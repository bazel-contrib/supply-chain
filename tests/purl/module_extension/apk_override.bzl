"""APK PURL type override used by module extension tests."""

load("@package_metadata//purl:providers.bzl", "PurlTypeInfo")

_APK_QUALIFIERS = {
    "arch": True,
    "distro": True,
    "upstream": True,
}

def _validate_apk(*, type, namespace, name, version, qualifiers, subpath):
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

def _normalize_apk(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type.lower(),
        namespace = namespace.lower() if namespace else namespace,
        name = name.lower() if name else name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )

def _apk_type_impl(ctx):
    return [
        PurlTypeInfo(
            validate = _validate_apk,
            normalize = _normalize_apk,
        ),
    ]

apk_type = rule(
    implementation = _apk_type_impl,
    provides = [
        PurlTypeInfo,
    ],
)
