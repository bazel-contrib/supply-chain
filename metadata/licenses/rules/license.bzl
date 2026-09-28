"""Declares rule `license`."""

load("//licenses/providers:license_kind_info.bzl", "LicenseKindInfo")
load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `license` targets.
KIND = "build.bazel.attribute.license"

def _license_impl(ctx):
    kind = ctx.attr.kind[LicenseKindInfo]
    attribute = {
        "kind": {
            "identifier": kind.identifier,
            "name": kind.name,
        },
        "label": str(ctx.label),
    }
    files = []

    if ctx.attr.text:
        attribute["text"] = ctx.file.text.path
        files.append(ctx.attr.text[DefaultInfo].files)

    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        output = output,
        content = json.encode(attribute),
    )

    return [
        DefaultInfo(
            files = depset(
                direct = [
                    output,
                ],
            ),
        ),
        PackageAttributeInfo(
            kind = KIND,
            attributes = output,
            files = files,
        ),
    ]

# Documentation lives on the public `license` macro below: Stardoc renders the
# macro, not this private rule.
_license = rule(
    implementation = _license_impl,
    attrs = {
        "kind": attr.label(
            mandatory = True,
            providers = [
                LicenseKindInfo,
            ],
        ),
        "text": attr.label(
            mandatory = False,
            allow_single_file = True,
        ),
    },
)

def license(
        # Disallow unnamed attributes.
        *,
        # `_license` attributes.
        name,
        kind,
        text = None,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring the license of a package or target.

    This is typically passed to the `attributes` of a `package_metadata` target.

    Usage:

    ```starlark
    load("@package_metadata//licenses/rules:license.bzl", "license")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            ":license",
        ],
        purl = "pkg:generic/example@1.0.0",
        visibility = ["//visibility:public"],
    )

    license(
        name = "license",
        kind = "@package_metadata//licenses/spdx:GPL-3.0-or-later",
        text = "COPYING",
    )
    ```

    A `license` target declares exactly one license kind. Since the
    `attributes` of a `package_metadata` target are keyed by `kind`, at most one
    `license` attribute per `package_metadata` target is meaningful; declare
    separate `package_metadata` targets for parts of a package that are under a
    different license. Packages available under a choice of licenses (e.g.,
    `MIT OR Apache-2.0`) and SPDX license exceptions are not modelled yet.

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.license`, and its `attributes` file holds a JSON
    object, encoded as UTF-8 and written without insignificant whitespace (the
    example below is indented for readability):

    ```json
    {
        "kind": {
            "identifier": "GPL-3.0-or-later",
            "name": "GNU General Public License v3.0 or later"
        },
        "label": "@@//:license",
        "text": "COPYING"
    }
    ```

    | Field | Description |
    | :---- | :---------- |
    | `kind.identifier` | The `identifier` of the `kind` of this license. |
    | `kind.name` | The `full_name` of the `kind` of this license. |
    | `label` | The label of this `license` target. |
    | `text` | The path of the license `text` file, relative to the execution root. Absent if no `text` was declared. |

    Args:
      name: A unique name for this target.
      kind: Required. The kind of license this license is classified as.

        This is typically a `license_kind` target. Targets for all SPDX
        licenses are predeclared in `@package_metadata//licenses/spdx`.
      text: The [File](https://bazel.build/rules/lib/builtins/File) with the
        text of the license.

        This is typically the `LICENSE` or `COPYING` file of the package. It is
        propagated to consumers so that it can be shipped alongside built
        artifacts.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    _license(
        # `_license` attributes.
        name = name,
        kind = kind,
        text = text,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        applicable_licenses = [],
    )
