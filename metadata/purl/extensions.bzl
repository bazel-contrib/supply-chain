"""Bzlmod extensions for PURL type customization."""

load("//purl/types:requirements.bzl", "TYPE_REQUIREMENTS")

visibility("public")

_REGISTRY_REPO_NAME = "package_metadata_purl_types"

_BUILTIN_PURL_TYPES = [
    "alpm",
    "apk",
    "bazel",
    "bitbucket",
    "bitnami",
    "cargo",
    "chrome-extension",
    "cocoapods",
    "composer",
    "conan",
    "conda",
    "cpan",
    "cran",
    "deb",
    "docker",
    "gem",
    "generic",
    "github",
    "golang",
    "hackage",
    "hex",
    "huggingface",
    "julia",
    "luarocks",
    "maven",
    "mlflow",
    "npm",
    "nuget",
    "oci",
    "opam",
    "otp",
    "pub",
    "pypi",
    "qpkg",
    "rpm",
    "swid",
    "swift",
    "vscode-extension",
    "yocto",
]
_NORMALIZER_FUNCTIONS = {
    "bitbucket": "normalize_bitbucket",
    "composer": "normalize_composer",
    "github": "normalize_github",
    "huggingface": "normalize_huggingface",
    "mlflow": "normalize_mlflow",
    "pypi": "normalize_pypi",
}

def _purl_type_registry_repo_impl(ctx):
    build_lines = [
        'package(default_visibility = ["//visibility:public"])',
        "",
        'exports_files(["validators.bzl", "normalizers.bzl"])',
        "",
    ]

    validators_lines = [
        '"""Generated PURL type validator registry."""',
        "",
    ]
    normalizers_lines = [
        '"""Generated PURL type normalizer registry."""',
        "",
    ]

    validator_loads = []
    validator_entries = []
    normalizer_loads = []
    normalizer_entries = []

    for type in sorted(ctx.attr.normalization_files.keys()):
        file = ctx.attr.normalization_files[type]
        function = ctx.attr.normalization_functions[type]
        statement = 'load("{}", "{}")'.format(str(file), function)
        if statement not in normalizer_loads:
            normalizer_loads.append(statement)
            normalizer_entries.append('    "{}": {},'.format(type, function))

    for type in sorted(ctx.attr.validation_files.keys()):
        file = ctx.attr.validation_files[type]
        function = ctx.attr.validation_functions[type]
        validator_loads.append('load("{}", "{}")'.format(file, function))
        validator_entries.append('    "{}": {},'.format(type, function))


    if validator_loads:
        validators_lines.extend(validator_loads)
        validators_lines.append("")
        validators_lines.append("TYPE_VALIDATORS = {")
        validators_lines.extend(validator_entries)
        validators_lines.append("}")
        validators_lines.append("")
    else:
        validators_lines.append("TYPE_VALIDATORS = {}")
        validators_lines.append("")

    if normalizer_loads:
        normalizers_lines.extend(normalizer_loads)
        normalizers_lines.append("")
        normalizers_lines.append("TYPE_NORMALIZERS = {")
        normalizers_lines.extend(normalizer_entries)
        normalizers_lines.append("}")
        normalizers_lines.append("")
    else:
        normalizers_lines.append("TYPE_NORMALIZERS = {}")
        normalizers_lines.append("")

    ctx.file("BUILD.bazel", "\n".join(build_lines))
    ctx.file("validators.bzl", "\n".join(validators_lines))
    ctx.file("normalizers.bzl", "\n".join(normalizers_lines))

_purl_type_registry_repo = repository_rule(
    implementation = _purl_type_registry_repo_impl,
    attrs = {
        "normalization_files": attr.string_keyed_label_dict(),
        "normalization_functions": attr.string_dict(),
        "validation_files": attr.string_keyed_label_dict(),
        "validation_functions": attr.string_dict(),
    },
)

def _purl_types_impl(ctx):
    normalization_files = {}
    normalization_functions = {}
    validation_files = {}
    validation_functions = {}

    for type in _BUILTIN_PURL_TYPES:
        normalization_files[type] = Label("//purl/types:{}.bzl".format(type.replace("-", "_")) if type in _NORMALIZER_FUNCTIONS else "//purl/types:helpers.bzl")
        normalization_functions[type] = _NORMALIZER_FUNCTIONS[type].replace("-", "_") if type in _NORMALIZER_FUNCTIONS else "identity_normalization"
        validation_files[type] = Label("//purl/types:{}.bzl".format(type.replace("-", "_")))
        validation_functions[type] = "validate_{}".format(type.replace("-", "_"))

    for module in ctx.modules:
        if module.is_root:
            continue
        for tag in module.tags.type:
            normalization_files[tag.name] = tag.normalize_file if tag.normalize_file != None else Label("//purl/types:helpers.bzl")
            normalization_functions[tag.name] = tag.normalize_function if tag.normalize_function != None else "identity_normalization"

            validation_files[tag.name] = tag.validation_file
            validation_functions[tag.name] = tag.validation_function if tag.validation_function != None else ""


    for module in ctx.modules:
        if not module.is_root:
            continue
        for tag in module.tags.type:
            normalization_files[tag.name] = tag.normalize_file if tag.normalize_file != None else Label("//purl/types:helpers.bzl")
            normalization_functions[tag.name] = tag.normalize_function if tag.normalize_function != None else "identity_normalization"

            validation_files[tag.name] = tag.validation_file
            validation_functions[tag.name] = tag.validation_function if tag.validation_function != None else ""

    _purl_type_registry_repo(
        name = _REGISTRY_REPO_NAME,
        normalization_files = normalization_files,
        normalization_functions = normalization_functions,
        validation_files = validation_files,
        validation_functions = validation_functions,
    )

purl_types = module_extension(
    implementation = _purl_types_impl,
    tag_classes = {
        "type": tag_class(
            attrs = {
                "name": attr.string(
                    doc = "PURL type name to register or override.",
                    mandatory = True,
                ),
                "normalize_file": attr.label(
                    doc = "File to load the normalizer function from.",
                    mandatory = False,
                ),
                "normalize_function": attr.string(
                    doc = "Normalizer function name to load from the specified file. Will use the identity normalizer if not specified.",
                    mandatory = False,
                ),
                "validation_file": attr.label(
                    doc = "File to load the validator function from.",
                    mandatory = True,
                ),
                "validation_function": attr.string(
                    doc = "Validator function name to load from the validation_file.",
                    mandatory = True,
                ),
            },
        ),
    },
)
