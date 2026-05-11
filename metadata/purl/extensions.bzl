"""Bzlmod extensions for PURL type customization."""

load("//purl:type_names.bzl", "BUILTIN_PURL_TYPES")

visibility("public")

_REGISTRY_REPO_NAME = "package_metadata_purl_types"

def _purl_type_registry_repo_impl(ctx):
    build_lines = [
        'package(default_visibility = ["//visibility:public"])',
        "",
    ]
    defs_lines = [
        '"""Generated PURL type override registry."""',
        "",
        "PURL_TYPE_OVERRIDES = {",
    ]
    validators_lines = [
        '"""Generated PURL type validator registry."""',
        "",
    ]

    validator_loads = []
    validator_entries = []

    for type in sorted(ctx.attr.registrations.keys()):
        target = ctx.attr.registrations[type]
        build_lines.extend([
            "alias(",
            '    name = "{}",'.format(type),
            '    actual = "{}",'.format(target),
            ")",
            "",
        ])
        defs_lines.append('    "{}": Label("@{}//:{}"),'.format(type, _REGISTRY_REPO_NAME, type))

        if target.startswith("@package_metadata//purl/types:"):
            function_name = type.replace("-", "_")
            file_name = function_name
            validator_loads.append('load("@package_metadata//purl/types:{}.bzl", "validate_{}")'.format(file_name, function_name))
            validator_entries.append('    "{}": validate_{},'.format(type, function_name))

    defs_lines.append("}")
    defs_lines.append("")

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

    ctx.file("BUILD.bazel", "\n".join(build_lines))
    ctx.file("defs.bzl", "\n".join(defs_lines))
    ctx.file("validators.bzl", "\n".join(validators_lines))

_purl_type_registry_repo = repository_rule(
    implementation = _purl_type_registry_repo_impl,
    attrs = {
        "registrations": attr.string_dict(
            doc = "Map of PURL type to target label providing PurlTypeInfo.",
        ),
    },
)

def _purl_types_impl(ctx):
    registrations = {
        type: "@package_metadata//purl/types:{}".format(type)
        for type in BUILTIN_PURL_TYPES
    }
    for module in ctx.modules:
        if module.is_root:
            continue
        for tag in module.tags.type:
            registrations[tag.name] = str(tag.target)

    for module in ctx.modules:
        if not module.is_root:
            continue
        for tag in module.tags.type:
            registrations[tag.name] = str(tag.target)

    _purl_type_registry_repo(
        name = _REGISTRY_REPO_NAME,
        registrations = registrations,
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
                "target": attr.label(
                    doc = "Target that provides PurlTypeInfo.",
                    mandatory = True,
                ),
            },
        ),
    },
)
