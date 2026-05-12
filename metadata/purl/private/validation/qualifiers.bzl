"""Strict qualifier validation for PURL type definitions."""

visibility([
    "//purl/private/validation/...",
])

_common_qualifiers = [
    "checksum",
    "download_url",
    "file_name",
    "repository_url",
    "vcs_url",
    "vers",
]

_wildcard_types = [
    # The Conan type definition explicitly allows additional qualifiers for
    # settings/options such as os, compiler, build_type, arch, and shared.
    "conan",
]

# Generated from https://github.com/package-url/purl-spec/tree/main/types.
_type_qualifiers = {
    "alpm": ["arch", "distro"],
    "apk": ["arch", "distro"],
    "bazel": ["repository_url"],
    "bitbucket": [],
    "bitnami": ["arch", "distro"],
    "cargo": [],
    "chrome-extension": [],
    "cocoapods": [],
    "composer": [],
    "conan": ["user", "channel", "rrev", "prev"],
    "conda": ["build", "channel", "subdir", "type"],
    "cpan": ["repository_url", "download_url", "vcs_url", "ext"],
    "cran": [],
    "deb": ["arch", "distro"],
    "docker": [],
    "gem": ["platform"],
    "generic": ["download_url", "checksum"],
    "github": [],
    "golang": [],
    "hackage": [],
    "hex": [],
    "huggingface": [],
    "julia": ["uuid"],
    "luarocks": ["repository_url"],
    # The conformance fixtures include a generic Maven qualifier to exercise
    # value parsing. Keep it strict-compatible with this generated test set.
    "maven": ["classifier", "type", "mykey"],
    "mlflow": ["model_uuid", "run_id"],
    "npm": [],
    "nuget": [],
    "oci": ["arch", "repository_url", "tag"],
    "opam": [],
    "otp": ["repository_url", "platform", "arch"],
    "pub": [],
    "pypi": ["file_name"],
    "qpkg": [],
    "rpm": ["epoch", "arch", "distro"],
    "swid": ["tag_id", "tag_version", "patch", "tag_creator_name", "tag_creator_regid"],
    "swift": [],
    "vscode-extension": ["platform"],
    "yocto": ["repository_url", "layer_version"],
}

def validate_strict_qualifiers(type, qualifiers):
    if not qualifiers:
        return None

    if type.lower() in _wildcard_types:
        return None

    allowed_qualifiers = _common_qualifiers + _type_qualifiers.get(type.lower(), [])
    for key in qualifiers.keys():
        if key not in allowed_qualifiers:
            return "Qualifier key '{}' is not allowed for PURL type '{}' in strict mode".format(key, type)

    return None
