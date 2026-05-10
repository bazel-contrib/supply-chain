"""Rule for exposing built-in PURL types as PurlTypeInfo targets."""

load("//purl:providers.bzl", "PurlTypeInfo")
load("//purl/private/normalization:normalization.bzl", "normalize")
load("//purl/private/validation:validation.bzl", "validate")

visibility([
    "//purl/private/...",
    "//purl/types:__pkg__",
])

def _validate_builtin(*, type, namespace, name, version, qualifiers, subpath):
    return validate(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )

def _normalize_builtin(*, type, namespace, name, version, qualifiers, subpath):
    purl, err = normalize(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
    if err:
        fail(err)
    return purl

def _builtin_type_impl(ctx):
    return [
        PurlTypeInfo(
            validate = _validate_builtin,
            normalize = _normalize_builtin,
        ),
    ]

builtin_type = rule(
    implementation = _builtin_type_impl,
    provides = [
        PurlTypeInfo,
    ],
)
