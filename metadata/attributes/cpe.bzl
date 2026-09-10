"""Declares rule `cpe`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

KIND = "build.bazel.attribute.cpe"

_DOC_STRING = """
Rule for declaring the CPE of a package or target.

This is typically a component of `package_metadata.attributes`.

- See https://en.wikipedia.org/wiki/Common_Platform_Enumeration
  for syntax.

Usage:

```starlark
load("@package_metadata//attributes:cpe.bzl", "cpe")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    purl = "...",
    attributes = [":cpe", ...],
)

cpe(
    name = "cpe",
    cpe = [
        "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*",
    ],
)
"""

def _cpe_impl(ctx):
    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        output = output,
        content = json.encode(ctx.attr.text),
    )
    return [
        DefaultInfo(
            files = depset(direct = [output]),
        ),
        PackageAttributeInfo(
            kind = KIND,
            attributes = output,
        ),
    ]

_cpe = rule(
    implementation = _cpe_impl,
    doc = _DOC_STRING.strip(),
    attrs = {
        "cpe": attr.string_list(
            mandatory = True,
            doc = """One or more CPE identifieres. UTF-8 encoding.""",
        ),
    },
)

def cpe(
        # Disallow unnamed attributes.
        *,
        # `_cpe` attributes.
        name,
        text = None,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    _cpe(
        # `_cpe` attributes.
        name = name,
        cpe = cpe,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        # This should be package_metadata, but we use the legacy name
        # to support bazel 7.
        applicable_licenses = [],
    )
