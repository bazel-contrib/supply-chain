"""Machine-readable PURL type definitions.

Derived from https://github.com/package-url/purl-spec/tree/main/types.
"""

visibility([
    "//purl/private/...",
])

TYPE_DEFINITIONS = {
    "alpm": {"namespace": "required", "lower_namespace": True, "lower_name": True, "qualifiers": ["arch"]},
    "apk": {"namespace": "required", "lower_namespace": True, "lower_name": True, "qualifiers": ["arch"]},
    "bazel": {"namespace": "prohibited"},
    "bitbucket": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "bitnami": {"namespace": "prohibited", "lower_name": True, "qualifiers": ["arch", "distro"]},
    "cargo": {"namespace": "prohibited"},
    "chrome-extension": {"namespace": "prohibited", "lower_name": True, "name_pattern": "chrome_extension_id", "version_pattern": "chrome_extension_version"},
    "cocoapods": {"namespace": "prohibited"},
    "composer": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "conan": {"qualifiers": ["arch", "build_type", "channel", "compiler", "compiler.runtime", "compiler.version", "os", "prev", "rrev", "shared", "user"]},
    "conda": {"namespace": "prohibited", "qualifiers": ["build", "channel", "subdir", "type"]},
    "cpan": {"namespace": "required", "cpan": True, "qualifiers": ["ext"]},
    "cran": {"namespace": "prohibited"},
    "deb": {"namespace": "required", "lower_namespace": True, "lower_name": True, "qualifiers": ["arch", "distro"]},
    "docker": {},
    "gem": {"namespace": "prohibited", "qualifiers": ["platform"]},
    "generic": {},
    "github": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "golang": {"namespace": "required", "lower_namespace": True, "lower_name": True},
    "hackage": {"namespace": "prohibited"},
    "hex": {"lower_namespace": True, "lower_name": True},
    "huggingface": {"lower_version": True},
    "julia": {"namespace": "prohibited", "qualifiers": ["uuid"], "required_qualifiers": ["uuid"]},
    "luarocks": {"lower_namespace": True, "lower_name": True},
    "maven": {"namespace": "required", "qualifiers": ["classifier", "type"]},
    "mlflow": {"namespace": "prohibited", "mlflow": True, "qualifiers": ["model_uuid", "run_id"]},
    "npm": {"lower_namespace": True, "lower_name": True},
    "nuget": {"namespace": "prohibited"},
    "oci": {"namespace": "prohibited", "lower_name": True, "lower_version": True, "qualifiers": ["arch", "tag"]},
    "opam": {"namespace": "prohibited"},
    "otp": {"namespace": "prohibited", "lower_name": True, "lower_subpath": True, "qualifiers": ["arch", "platform"]},
    "pub": {"namespace": "prohibited", "lower_name": True, "name_pattern": "pub_name"},
    "pypi": {"namespace": "prohibited", "lower_name": True, "lower_version": True, "pypi": True},
    "qpkg": {"namespace": "required", "lower_namespace": True},
    "rpm": {"namespace": "required", "lower_namespace": True, "qualifiers": ["arch", "distro", "epoch"]},
    "swid": {"qualifiers": ["patch", "tag_creator_name", "tag_creator_regid", "tag_id", "tag_version"], "required_qualifiers": ["tag_id"]},
    "swift": {"namespace": "required"},
    "vscode-extension": {"namespace": "required", "lower_namespace": True, "lower_name": True, "lower_version": True, "qualifiers": ["platform"]},
    "yocto": {"lower_namespace": True, "qualifiers": ["layer_version"]},
}
