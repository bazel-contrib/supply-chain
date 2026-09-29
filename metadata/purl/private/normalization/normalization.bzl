"""Utils to normalize [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private/normalization:alpm.bzl", "normalize_alpm")
load("//purl/private/normalization:bitbucket.bzl", "normalize_bitbucket")
load("//purl/private/normalization:composer.bzl", "normalize_composer")
load("//purl/private/normalization:github.bzl", "normalize_github")
load("//purl/private/normalization:hackage.bzl", "normalize_hackage")
load("//purl/private/normalization:huggingface.bzl", "normalize_huggingface")
load("//purl/private/normalization:mlflow.bzl", "normalize_mlflow")
load("//purl/private/normalization:pub.bzl", "normalize_pub")
load("//purl/private/normalization:pypi.bzl", "normalize_pypi")

visibility([
    "//purl/private",
])

_normalizers = {
    "alpm": normalize_alpm,
    "bitbucket": normalize_bitbucket,
    "composer": normalize_composer,
    "github": normalize_github,
    "hackage": normalize_hackage,
    "huggingface": normalize_huggingface,
    "mlflow": normalize_mlflow,
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
        subpath = None,
        check = True):
    if not type:
        return None, "Mandatory property 'type' not set"

    components = {
        "name": name,
        "namespace": namespace,
        "qualifiers": qualifiers,
        "subpath": subpath,
        "type": type.lower(),
        "version": version,
    }

    normalizer = _normalizers.get(components["type"])
    if check and normalizer:
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
