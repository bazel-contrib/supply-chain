"""Declares rule `fake_package_attribute`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

def _fake_package_attribute_impl(ctx):
    attributes = ctx.actions.declare_file("{}.json".format(ctx.attr.name))

    ctx.actions.write(
        output = attributes,
        content = json.encode({
            "source": str(ctx.label),
            "value": ctx.attr.value,
        }),
    )

    return [
        PackageAttributeInfo(
            kind = ctx.attr.kind,
            attributes = attributes,
        ),
    ]

_fake_package_attribute = rule(
    implementation = _fake_package_attribute_impl,
    attrs = {
        "kind": attr.string(
            mandatory = True,
        ),
        "value": attr.string(
            mandatory = True,
        ),
    },
    provides = [
        PackageAttributeInfo,
    ],
    doc = """
""".strip(),
)

def fake_package_attribute(
        # Disallow unnamed attributes.
        *,
        # `_fake_package_attribute` attributes.
        name,
        kind,
        value,
        # Common attributes (subset since this target is non-configurable).
        visibility = None):
    _fake_package_attribute(
        # `_fake_package_attribute` attributes.
        name = name,
        kind = kind,
        value = value,

        # Common attributes.
        visibility = visibility,
        applicable_licenses = [],
    )
