"""Root-module APKO PURL type override."""

load("@package_metadata//purl:providers.bzl", "PurlTypeInfo")

def _validate_root_apko(*, type, namespace, name, version, qualifiers, subpath):
    if type != "apko":
        return "Expected apko PURL type, got {}".format(type)
    if not name:
        return "Mandatory property 'name' not set"
    return None

def _normalize_root_apko(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = type.lower(),
        namespace = namespace.lower() if namespace else namespace,
        name = name.lower() if name else name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )

def _root_apko_type_impl(ctx):
    return [
        PurlTypeInfo(
            validate = _validate_root_apko,
            normalize = _normalize_root_apko,
        ),
    ]

root_apko_type = rule(
    implementation = _root_apko_type_impl,
    provides = [
        PurlTypeInfo,
    ],
)
