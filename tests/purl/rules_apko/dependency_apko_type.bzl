"""Dependency-provided APKO PURL type registration."""

load("@package_metadata//purl:providers.bzl", "PurlTypeInfo")

def _validate_dependency_apko(*, type, namespace, name, version, qualifiers, subpath):
    return "dependency apko validator"

def _normalize_dependency_apko(*, type, namespace, name, version, qualifiers, subpath):
    return struct(
        type = "dependency-apko",
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )

def _dependency_apko_type_impl(ctx):
    return [
        PurlTypeInfo(
            validate = _validate_dependency_apko,
            normalize = _normalize_dependency_apko,
        ),
    ]

dependency_apko_type = rule(
    implementation = _dependency_apko_type_impl,
    provides = [
        PurlTypeInfo,
    ],
)
