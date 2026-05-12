"""Utils to normalize [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private/normalization:alpm.bzl", "normalize_alpm")
load("//purl/private/normalization:hackage.bzl", "normalize_hackage")
load("//purl/private/normalization:pub.bzl", "normalize_pub")
load("//purl/private/normalization:pypi.bzl", "normalize_pypi")

visibility([
    "//purl/private",
])

_normalizers = {
    "alpm": normalize_alpm,
    "hackage": normalize_hackage,
    "pub": normalize_pub,
    "pypi": normalize_pypi,
}

def normalize(
        *,
        type = None,
        namespace = None,
        name = None,
        version = None,
        qualifiers = {},
        subpath = None):
    if not type:
        return None, "Mandatory property 'type' not set"

    components = {
        "type": type.lower(),
        "namespace": namespace,
        "name": name,
        "version": version,
        "qualifiers": qualifiers,
        "subpath": subpath,
    }

    normalizer = _normalizers.get(components["type"])
    if normalizer:
        components = normalizer(components)

    purl = struct(
        type = components["type"],
        namespace = [segment for segment in components["namespace"].split("/") if segment] if components["namespace"] else None,
        name = components["name"],
        version = components["version"],
        qualifiers = components["qualifiers"],
        subpath = [segment for segment in components["subpath"].split("/") if segment] if components["subpath"] else None,
    )
    return purl, None
