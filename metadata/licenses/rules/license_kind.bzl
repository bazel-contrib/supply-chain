"""Declares rule `license_kind`."""

load("//licenses/providers:license_kind_info.bzl", "LicenseKindInfo")
load("//licenses/spdx:licenses.bzl", "LICENSES")

visibility("public")

_IDSTRING_CHARS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-"

def _check_identifier(identifier):
    """Fails unless the identifier is an SPDX identifier or a valid custom one.

    A custom identifier is an SPDX `idstring`, optionally prefixed by
    `LicenseRef-`.
    """
    if LICENSES.get(identifier.lower(), {}).get("identifier") == identifier:
        return
    idstring = identifier
    if identifier.startswith("AdditionRef-"):
        fail("invalid license identifier '{}': custom exceptions must be declared with `license_exception`".format(identifier))
    if identifier.startswith("LicenseRef-"):
        idstring = identifier[len("LicenseRef-"):]
    invalid = sorted({c: True for c in idstring.elems() if c not in _IDSTRING_CHARS})
    if not idstring or invalid:
        fail("invalid license identifier '{}': a custom identifier may only contain letters, digits, '.' and '-' (SPDX idstring){}".format(
            identifier,
            ", found " + ", ".join(["'{}'".format(c) for c in invalid]) if invalid else "",
        ))

def _license_kind_impl(ctx):
    _check_identifier(ctx.attr.identifier)
    return [
        LicenseKindInfo(
            identifier = ctx.attr.identifier,
            name = ctx.attr.full_name,
            deprecation = ctx.attr.deprecation,
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
        deprecation = None,
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
    are declared with the separate `license_exception` rule.

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
        the license, but may also be a custom value (e.g., in case of a
        commercial license). A custom value may be written
        `LicenseRef-<idstring>`; custom exceptions use `AdditionRef-<idstring>`
        and are declared with `license_exception`. A value that is not on the
        SPDX list may only contain letters, digits, `.` and `-` (the SPDX
        `idstring` characters), after the optional prefix; any other character
        fails the build.

        This identifies a single license; it is not an SPDX license expression
        (e.g., `MIT OR Apache-2.0`, or `GPL-2.0-only WITH
        Classpath-exception-2.0`).
      full_name: Required. The (human readable) name of the license (e.g.,
        `Apache License 2.0`, `European Union Public License 1.1`).
      deprecation: An optional message explaining why this target is deprecated.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    _license_kind(
        # `_license_kind` attributes.
        name = name,
        identifier = identifier,
        full_name = full_name,

        # Common attributes.
        deprecation = deprecation,
        tags = tags,
        visibility = visibility,
        applicable_licenses = [],
    )
