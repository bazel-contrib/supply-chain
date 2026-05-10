"""Machine-readable PURL type definitions.

Derived from https://github.com/package-url/purl-spec/tree/main/types.
"""

visibility([
    "//purl/private/...",
])

TYPE_DEFINITIONS = {
    "alpm": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "apk": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "bazel": {"namespace": "prohibited"},
    "bitbucket": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "bitnami": {"namespace": "prohibited", "lower_name": True},
    "cargo": {"namespace": "prohibited"},
    "chrome-extension": {"namespace": "prohibited", "lower_name": True, "name_pattern": "chrome_extension_id", "version_pattern": "chrome_extension_version"},
    "cocoapods": {"namespace": "prohibited"},
    "composer": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "conan": {},
    "conda": {"namespace": "prohibited"},
    "cpan": {"namespace": "required", "cpan": True},
    "cran": {"namespace": "prohibited"},
    "deb": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "docker": {},
    "gem": {"namespace": "prohibited"},
    "generic": {},
    "github": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "golang": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "hackage": {"namespace": "prohibited"},
    "hex": {"lower_namespace": True, "lower_name": True},
    "huggingface": {"lower_version": True},
    "julia": {"namespace": "prohibited", "required_qualifiers": ["uuid"]},
    "luarocks": {"lower_namespace": True, "lower_name": True},
    "maven": {"namespace": "required"},
    "mlflow": {"namespace": "prohibited", "mlflow": True},
    "npm": {"lower_namespace": True, "lower_name": True},
    "nuget": {"namespace": "prohibited"},
    "oci": {"namespace": "prohibited", "lower_name": True, "lower_version": True},
    "opam": {"namespace": "prohibited"},
    "otp": {"namespace": "prohibited", "lower_name": True, "lower_subpath": True},
    "pub": {"namespace": "prohibited", "lower_name": True, "name_pattern": "pub_name"},
    "pypi": {"namespace": "prohibited", "lower_name": True, "lower_version": True, "pypi": True},
    "qpkg": {"namespace": "required", "lower_namespace": True},
    "rpm": {"namespace": "required", "lower_namespace": True},
    "swid": {"required_qualifiers": ["tag_id"]},
    "swift": {"namespace": "required"},
    "vscode-extension": {"namespace": "required", "lower_namespace": True, "lower_name": True, "lower_version": True},
    "yocto": {"lower_namespace": True},
}
