"""Declares rule `copyright_notice`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `copyright_notice`
# targets.
KIND = "build.bazel.attribute.copyright_notice"

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

# Documentation lives on the public `copyright_notice` macro below: Stardoc
# renders the macro, not this private rule.
_copyright_notice = rule(
    implementation = _copyright_notice_impl,
    attrs = {
        "text": attr.string_list(
            mandatory = True,
        ),
    },
)

def copyright_notice(
        # Disallow unnamed attributes.
        *,
        # `_copyright_notice` attributes.
        name,
        text,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring the copyright notice of a package or target.

    This is typically passed to the `attributes` of a `package_metadata` target.

    Copyright `text` entries are assumed to be UTF-8 encoded. Tools for building
    SBOMs presume that.

    Usage:

    ```starlark
    load("@package_metadata//attributes:copyright_notice.bzl", "copyright_notice")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            ":copyright_notice",
        ],
        purl = "pkg:generic/example@1.0.0",
        visibility = ["//visibility:public"],
    )

    copyright_notice(
        name = "copyright_notice",
        text = [
            "Copyright © 2026 The authors of this stuff.",
            "Copyright © 1999 A. Cöntributor",
        ],
    )
    ```

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.copyright_notice`, and its `attributes` file holds
    the `text` entries as a JSON array of strings, encoded as UTF-8 and written
    without insignificant whitespace:

    ```json
    [
        "Copyright © 2026 The authors of this stuff.",
        "Copyright © 1999 A. Cöntributor"
    ]
    ```

    Args:
      name: A unique name for this target.
      text: Required. One or more copyright notices. UTF-8 encoding.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

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
