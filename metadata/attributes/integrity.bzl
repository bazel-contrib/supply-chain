"""Declares rule `integrity`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `integrity` targets.
KIND = "build.bazel.attribute.integrity"

# Digest sizes, in bytes, of the algorithms tools are known to report.
_DIGEST_SIZES = {
    "sha1": 20,
    "sha256": 32,
    "sha384": 48,
    "sha512": 64,
}

_BASE64_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/-_"

def _validate(sri):
    """Fails if `sri` is not `<algorithm>-<base64 digest>` or its digest has the wrong size."""
    algorithm, separator, digest = sri.partition("-")
    if not algorithm or not separator:
        fail("integrity {} is not of the form <algorithm>-<base64 digest>".format(repr(sri)))

    # An SRI string can carry options after the digest.
    digest = digest.partition("?")[0].rstrip("=")
    if not digest:
        fail("integrity {} has no digest".format(repr(sri)))
    for char in digest.elems():
        if char not in _BASE64_ALPHABET:
            fail("integrity {}: the digest is not base64".format(repr(sri)))

    # Algorithms that are not known are passed through, and tools may drop them.
    size = _DIGEST_SIZES.get(algorithm.lower())
    if size != None and len(digest) * 6 // 8 != size:
        fail("integrity {}: a {} digest has {} bytes".format(repr(sri), algorithm, size))

def _integrity_impl(ctx):
    if not ctx.attr.values:
        fail("values must contain at least one entry")
    for value in ctx.attr.values:
        _validate(value)

    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        output = output,
        content = json.encode(ctx.attr.values),
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

# Documentation lives on the public `integrity` macro below: Stardoc renders the
# macro, not this private rule.
_integrity = rule(
    implementation = _integrity_impl,
    attrs = {
        "values": attr.string_list(
            mandatory = True,
        ),
    },
)

def integrity(
        # Disallow unnamed attributes.
        *,
        # `_integrity` attributes.
        name,
        values,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring the integrity of a package or target.

    This is typically passed to the `attributes` of a `package_metadata` target.

    Each integrity is a Subresource Integrity (SRI) string: the hash algorithm
    name, a dash, then the base64-encoded digest, e.g. `sha256-<base64 digest>`.
    The form and base64 encoding of each value are validated, and so is the size
    of the digest of `sha1`, `sha256`, `sha384` and `sha512`. Values of other
    algorithms are passed through as-is: tools consuming the attribute map them
    to the format they produce, and may drop algorithms they do not support.

    Entries are assumed to be UTF-8 encoded. Tools for building SBOMs presume
    that.

    Usage:

    ```starlark
    load("@package_metadata//attributes:integrity.bzl", "integrity")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            ":integrity",
        ],
        purl = "pkg:maven/org.apache.commons/commons-lang3@3.12.0",
        visibility = ["//visibility:public"],
    )

    integrity(
        name = "integrity",
        values = [
            "sha1-he0IF6+DokrY2mjCtQlN5pgzmDw=",
            "sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q=",
        ],
    )
    ```

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.integrity`, and its `attributes` file holds the
    `values` entries as a JSON array of strings, encoded as UTF-8 and
    written without insignificant whitespace:

    ```json
    [
        "sha1-he0IF6+DokrY2mjCtQlN5pgzmDw=",
        "sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="
    ]
    ```

    Args:
      name: A unique name for this target.
      values: Required. A list of SRI strings (`<algorithm>-<base64 digest>`)
        for the package, e.g. `["sha256-<base64 digest>"]`. Must contain at
        least one entry. UTF-8 encoding.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    if not values:
        fail("values must contain at least one entry")

    _integrity(
        # `_integrity` attributes.
        name = name,
        values = values,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        # This should be package_metadata, but we use the legacy name
        # to support bazel 7.
        applicable_licenses = [],
    )
