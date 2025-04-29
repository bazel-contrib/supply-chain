"""Declares rule `package_attribute_override`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")
load("//providers:package_attribute_override_info.bzl", "PackageAttributeOverrideInfo")

visibility("public")

def _package_attribute_override_impl(ctx):
    return [
        PackageAttributeOverrideInfo(
            package = Label(ctx.attr.package),
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

def _macro_impl(package, **kwargs):
    _package_attribute_override(
        package = str(package),
        applicable_licenses = [],
        **kwargs
    )

package_attribute_override = macro(
    implementation = _macro_impl,
    inherit_attrs = _package_attribute_override,
    attrs = {
        # This trick of defining `package` as `string` in rule
        # `package_attribute_override` above and as `label` in this macro allows
        # us to emulate no-dep labels.
        "package": attr.label(
            mandatory = True,
            configurable = False,
            doc = """
The label of the package these overrides are for.
""".strip(),
        ),
    },
)
