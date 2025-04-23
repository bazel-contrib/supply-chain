"""Declares rule `package_attribute_override`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")
load("//providers:package_attribute_override_info.bzl", "PackageAttributeOverrideInfo")

visibility("public")

def _package_attribute_override_impl(ctx):
    return [
        PackageAttributeOverrideInfo(
            package = ctx.label.relative(ctx.attr.package),
            overrides = [o[PackageAttributeInfo] for o in ctx.attr.overrides],
        ),
    ]

_package_attribute_override = rule(
    implementation = _package_attribute_override_impl,
    attrs = {
        "overrides": attr.label_list(
            mandatory = False,
            providers = [
                PackageAttributeInfo,
            ],
        ),
        "package": attr.string(
            mandatory = True,
            doc = """
The label of the package these overrides are for.
""".strip(),
        ),
    },
    provides = [
        PackageAttributeOverrideInfo,
    ],
    doc = """
Rule for declaring overrides for a `package_metadata` attribute.
""".strip(),
)

def package_attribute_override(
        # Disallow unnamed attributes.
        *,
        # `_package_attribute_override` attributes.
        name,
        package,
        overrides = [],
        # Common attributes (subset since this target is non-configurable).
        visibility = None):
    _package_attribute_override(
        # `_package_attribute_override` attributes.
        name = name,
        package = package,
        overrides = overrides,

        # Common attributes.
        visibility = visibility,
        applicable_licenses = [],
    )
