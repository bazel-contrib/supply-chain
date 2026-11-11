"""Declares rule `license_exception`."""

load("//licenses/providers:license_exception_info.bzl", "LicenseExceptionInfo")

visibility("public")

def _license_exception_impl(ctx):
    return [
        LicenseExceptionInfo(
            identifier = ctx.attr.identifier,
            name = ctx.attr.full_name,
            deprecation = ctx.attr.deprecation,
        ),
    ]

_license_exception = rule(
    implementation = _license_exception_impl,
    attrs = {
        "full_name": attr.string(
            mandatory = True,
        ),
        "identifier": attr.string(
            mandatory = True,
        ),
    },
    provides = [
        LicenseExceptionInfo,
    ],
)

def license_exception(
        # Disallow unnamed attributes.
        *,
        # `_license_exception` attributes.
        name,
        identifier,
        full_name,
        # Common attributes (subset since this target is non-configurable).
        deprecation = None,
        tags = None,
        visibility = None):
    """Rule for declaring `LicenseExceptionInfo`.

    A `license_exception` identifies an SPDX license exception, which is only
    valid with a license in a `WITH` expression and is never a standalone
    license.

    Args:
      name: A unique name for this target.
      identifier: Required. The unique identifier of the exception (e.g.,
        `Classpath-exception-2.0`).
      full_name: Required. The (human readable) name of the exception.
      deprecation: An optional message explaining why this target is deprecated.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    _license_exception(
        # `_license_exception` attributes.
        name = name,
        identifier = identifier,
        full_name = full_name,

        # Common attributes.
        deprecation = deprecation,
        tags = tags,
        visibility = visibility,
        applicable_licenses = [],
    )
