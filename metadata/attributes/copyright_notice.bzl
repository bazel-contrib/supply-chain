"""Declares rule `copyright_notice`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

KIND = "build.bazel.attribute.copyright_notice"

_DOC_STRING = """
Rule for declaring the copyright_notice of a package or target.

This is typically a component of `package_metadata.attributes`.

- Copyright `text` fields are assumeed to be UTF-8 encoded.
  Tools for building SBOMs presume that.

Usage:

```starlark
load("@package_metadata//attributes:copyright_notice.bzl", "copyright_notice")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    purl = "...",
    attributes = [
        ":copyright_notice", ...
    ],
    visibility = ["//visibility:public"],
)

copyright_notice(
    name = "copyright_notice",
    text = [
        "Copyright © 2026 The authors of this stuff.",
        "Copyright © 1999 A. Cöntributor",
    ],
)
"""

def _copyright_notice_impl(ctx):
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

_copyright_notice = rule(
    implementation = _copyright_notice_impl,
    doc = _DOC_STRING.strip(),
    attrs = {
        "text": attr.string_list(
            mandatory = True,
            doc = """One or more copyright notices. UTF-8 encoding.""",
        ),
    },
)

def copyright_notice(
        # Disallow unnamed attributes.
        *,
        # `_copyright_notice` attributes.
        name,
        text = None,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    _copyright_notice(
        # `_copyright_notice` attributes.
        name = name,
        text = text,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        # This should be package_metadata, but we use the legacy name
        # to support bazel 7.
        applicable_licenses = [],
    )
