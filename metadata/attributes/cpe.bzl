"""Declares rule `cpe`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `cpe` targets.
KIND = "build.bazel.attribute.cpe"

def _cpe_impl(ctx):
    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        output = output,
        content = json.encode(ctx.attr.identifiers),
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

# Documentation lives on the public `cpe` macro below: Stardoc renders the
# macro, not this private rule.
_cpe = rule(
    implementation = _cpe_impl,
    attrs = {
        "identifiers": attr.string_list(
            mandatory = True,
        ),
    },
)

def cpe(
        # Disallow unnamed attributes.
        *,
        # `_cpe` attributes.
        name,
        identifiers,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring the CPEs of a package or target.

    This is typically passed to the `attributes` of a `package_metadata` target.

    `identifiers` entries are assumed to be UTF-8 encoded. Tools for building SBOMs
    presume that.

    Usage:

    ```starlark
    load("@package_metadata//attributes:cpe.bzl", "cpe")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            ":cpe",
        ],
        purl = "pkg:generic/ntp@4.2.8",
        visibility = ["//visibility:public"],
    )

    cpe(
        name = "cpe",
        identifiers = [
            "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*",
        ],
    )
    ```

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.cpe`, and its `attributes` file holds the `identifiers`
    entries as a JSON array of strings, encoded as UTF-8 and written without
    insignificant whitespace:

    ```json
    [
        "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*"
    ]
    ```

    Args:
      name: A unique name for this target.
      identifiers: Required. One or more
        [CPE](https://en.wikipedia.org/wiki/Common_Platform_Enumeration)
        identifiers of the package. UTF-8 encoding.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    _cpe(
        # `_cpe` attributes.
        name = name,
        identifiers = identifiers,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        # This should be package_metadata, but we use the legacy name
        # to support bazel 7.
        applicable_licenses = [],
    )
