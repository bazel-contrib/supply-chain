"""Declares rule `package_metadata`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")
load("//providers:package_metadata_info.bzl", "PackageMetadataInfo")

visibility("public")

def _resolve_attributes(*, label, attributes, toolchain):
    attribute_map = {a.kind: a for a in attributes}

    if toolchain:
        overrides = toolchain.package_metadata.overrides.get(label, None)
        if overrides:
            for kind, attribute in overrides.overrides.items():
                attribute_map[kind] = attribute

    metadata = {kind: attribute.attributes.path for kind, attribute in attribute_map.items()}
    return metadata, [attribute.files for attribute in attribute_map.values()]

def _create_package_metadata(*, actions, output, label, purl, attributes = [], toolchain = None):
    resolved_attributes, files = _resolve_attributes(
        label = label,
        attributes = attributes,
        toolchain = toolchain,
    )

    actions.write(
        output = output,
        content = json.encode({
            "attributes": resolved_attributes,
            "label": str(label),
            "purl": purl,
        }),
    )

    return PackageMetadataInfo(
        metadata = output,
        files = files,
    )

def _package_metadata_impl(ctx):
    attributes = [a[PackageAttributeInfo] for a in ctx.attr.attributes]

    metadata = ctx.actions.declare_file("{}.package-metadata.json".format(ctx.attr.name))
    info = _create_package_metadata(
        actions = ctx.actions,
        output = metadata,
        label = ctx.label,
        purl = ctx.attr.purl,
        attributes = [a[PackageAttributeInfo] for a in ctx.attr.attributes],
        toolchain = ctx.toolchains["//toolchains:type"],
    )

    return [
        DefaultInfo(
            files = info.files,
        ),
        info,
    ]

_package_metadata = rule(
    implementation = _package_metadata_impl,
    attrs = {
        "attributes": attr.label_list(
            mandatory = False,
            doc = """
A list of `attributes` of the package (e.g., source location, license, ...).
""".strip(),
            providers = [
                PackageAttributeInfo,
            ],
        ),
        "purl": attr.string(
            mandatory = True,
            doc = """
The [PURL](https://github.com/package-url/purl-spec) uniquely identifying this
package.
""".strip(),
        ),
    },
    provides = [
        PackageMetadataInfo,
    ],
    doc = """
Rule for declaring `PackageMetadataInfo`, typically of a `bzlmod` module.
""".strip(),
    toolchains = [
        config_common.toolchain_type("//toolchains:type", mandatory = False),
    ],
)

def package_metadata(
        # Disallow unnamed attributes.
        *,
        # `_package_metadata` attributes.
        name,
        purl,
        attributes = [],
        # Common attributes (subset since this target is non-configurable).
        visibility = None):
    _package_metadata(
        # `_package_metadata` attributes.
        name = name,
        purl = purl,
        attributes = attributes,

        # Common attributes.
        visibility = visibility,
        applicable_licenses = [],
    )
