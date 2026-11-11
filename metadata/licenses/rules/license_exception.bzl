"""Declares rule `license_exception`."""

load("//licenses/providers:license_exception_info.bzl", "LicenseExceptionInfo")
load("//licenses/spdx:licenses.bzl", "EXCEPTIONS")

visibility("public")

_IDSTRING_CHARS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-"

def _check_identifier(identifier):
    if EXCEPTIONS.get(identifier.lower(), {}).get("identifier") == identifier:
        return
    prefix = "AdditionRef-"
    if not identifier.startswith(prefix):
        fail("invalid license exception identifier '{}': a custom exception must start with 'AdditionRef-'".format(identifier))
    idstring = identifier[len(prefix):]
    invalid = sorted({c: True for c in idstring.elems() if c not in _IDSTRING_CHARS})
    if not idstring or invalid:
        fail("invalid license exception identifier '{}': a custom identifier may only contain letters, digits, '.' and '-' (SPDX idstring){}".format(
            identifier,
            ", found " + ", ".join(["'{}'".format(c) for c in invalid]) if invalid else "",
        ))

def _license_exception_impl(ctx):
    _check_identifier(ctx.attr.identifier)
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
        `Classpath-exception-2.0`). Custom exceptions must use an
        `AdditionRef-<idstring>` identifier.
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
