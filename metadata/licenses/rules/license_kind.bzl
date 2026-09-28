"""Declares rule `license_kind`."""

load("//licenses/providers:license_kind_info.bzl", "LicenseKindInfo")

visibility("public")

def _license_kind_impl(ctx):
    return [
        LicenseKindInfo(
            identifier = ctx.attr.identifier,
            name = ctx.attr.full_name,
        ),
    ]

# Documentation lives on the public `license_kind` macro below: Stardoc renders
# the macro, not this private rule.
_license_kind = rule(
    implementation = _license_kind_impl,
    attrs = {
        "full_name": attr.string(
            mandatory = True,
        ),
        "identifier": attr.string(
            mandatory = True,
        ),
    },
    provides = [
        LicenseKindInfo,
    ],
)

def license_kind(
        # Disallow unnamed attributes.
        *,
        # `_license_kind` attributes.
        name,
        identifier,
        full_name,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring `LicenseKindInfo`.

    A `license_kind` identifies a license; a `license` target refers to it to
    declare that a package is available under that license.

    Targets for all [SPDX licenses](https://spdx.org/licenses/) are predeclared
    in `@package_metadata//licenses/spdx`, so declaring a `license_kind`
    yourself is only necessary for licenses that are not in the SPDX list (e.g.,
    a commercial license).

    [SPDX license exceptions](https://spdx.org/licenses/exceptions-index.html)
    (e.g., `Classpath-exception-2.0`) are a separate list and are **not**
    predeclared.

    Usage:

    ```starlark
    load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")

    license_kind(
        name = "acme_commercial",
        full_name = "ACME Commercial License 2.0",
        identifier = "LicenseRef-acme-commercial-2.0",
        visibility = ["//visibility:public"],
    )
    ```

    Args:
      name: A unique name for this target.
      identifier: Required. The unique identifier of the license (e.g.,
        `Apache-2.0`, `EUPL-1.1`).

        This is typically the [SPDX identifier](https://spdx.org/licenses/) of
        the license, but may also be a non-standard value (e.g., in case of a
        commercial license). It is not validated.

        This identifies a single license; it is not an SPDX license expression
        (e.g., `MIT OR Apache-2.0`, or `GPL-2.0-only WITH
        Classpath-exception-2.0`).
      full_name: Required. The (human readable) name of the license (e.g.,
        `Apache License 2.0`, `European Union Public License 1.1`).
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    _license_kind(
        # `_license_kind` attributes.
        name = name,
        identifier = identifier,
        full_name = full_name,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        applicable_licenses = [],
    )
