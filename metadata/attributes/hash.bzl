"""Declares rule `hash`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `hash` targets.
KIND = "build.bazel.attribute.hash"

def _hash_impl(ctx):
    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    entries = [
        {"alg": alg, "content": ctx.attr.hashes[alg]}
        for alg in sorted(ctx.attr.hashes)
    ]
    ctx.actions.write(
        output = output,
        content = json.encode(entries),
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

# Documentation lives on the public `hash` macro below: Stardoc renders the
# macro, not this private rule.
_hash = rule(
    implementation = _hash_impl,
    attrs = {
        "hashes": attr.string_dict(
            mandatory = True,
        ),
    },
)

def hash(
        # Disallow unnamed attributes.
        *,
        # `_hash` attributes.
        name,
        hashes,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring the content hashes of a package or target.

    This is typically passed to the `attributes` of a `package_metadata` target.

    `hashes` maps a CycloneDX hash-algorithm name to its hex digest. Algorithm
    names use the CycloneDX `hash-alg` spelling (`MD5`, `SHA-1`, `SHA-256`,
    `SHA-512`, ...). Tools for building SBOMs presume that spelling.

    Usage:

    ```starlark
    load("@package_metadata//attributes:hash.bzl", "hash")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            ":hash",
        ],
        purl = "pkg:maven/antlr/antlr@2.7.7",
        visibility = ["//visibility:public"],
    )

    hash(
        name = "hash",
        hashes = {
            "SHA-256": "88fbda4b912596b9f56e8e12e580cc954bacfb51776ecfddd3e18fc1cf56dc4c",
        },
    )
    ```

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.hash`, and its `attributes` file holds a JSON array of
    `{alg, content}` objects sorted by `alg`, written without insignificant
    whitespace:

    ```json
    [
        {"alg": "SHA-256", "content": "88fbda4b...dc4c"}
    ]
    ```

    Args:
      name: A unique name for this target.
      hashes: Required. A map of CycloneDX hash-algorithm name to hex digest.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    _hash(
        # `_hash` attributes.
        name = name,
        hashes = hashes,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        # This should be package_metadata, but we use the legacy name
        # to support bazel 7.
        applicable_licenses = [],
    )
