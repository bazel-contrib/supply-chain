"""Declares rule `patch_set`."""

load("//attributes:patch.bzl", "PatchInfo")
load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `patch_set` targets.
KIND = "build.bazel.attribute.patch_set"

def _patch_value(info):
    value = {
        "description": info.description,
        "integrity": info.integrity,
        "type": info.type,
    }
    if info.url:
        value["url"] = info.url
    if info.file:
        value["file"] = info.file
    if info.resolves:
        value["resolves"] = info.resolves
    return value

def _patch_set_impl(ctx):
    if not ctx.attr.patches:
        fail("patches must contain at least one entry")

    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        output = output,
        content = json.encode([_patch_value(p[PatchInfo]) for p in ctx.attr.patches]),
    )
    return [
        DefaultInfo(
            files = depset(direct = [output]),
        ),
        PackageAttributeInfo(
            kind = KIND,
            attributes = output,
            files = [p[PackageAttributeInfo].files for p in ctx.attr.patches],
        ),
    ]

# Documentation lives on the public `patch_set` macro below: Stardoc renders the
# macro, not this private rule.
_patch_set = rule(
    implementation = _patch_set_impl,
    attrs = {
        "patches": attr.label_list(
            mandatory = True,
            providers = [PatchInfo, PackageAttributeInfo],
        ),
    },
)

def patch_set(
        # Disallow unnamed attributes.
        *,
        # `_patch_set` attributes.
        name,
        patches,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring the set of patches applied to a package or target.

    This is typically passed to the `attributes` of a `package_metadata` target.
    Each entry of `patches` is a `patch` target.

    Usage:

    ```starlark
    load("@package_metadata//attributes:patch.bzl", "patch")
    load("@package_metadata//attributes:patch_set.bzl", "patch_set")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            ":patches",
        ],
        purl = "pkg:generic/ntp@4.2.8",
        visibility = ["//visibility:public"],
    )

    patch_set(
        name = "patches",
        patches = [":fix_overflow"],
    )

    patch(
        name = "fix_overflow",
        description = "Fixes a buffer overflow in the parser.",
        integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
        type = "backport",
        url = "https://example.com/patches/fix_overflow.patch",
    )
    ```

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.patch_set`, and its `attributes` file holds a JSON
    array of the objects described in `patch`, in the order of `patches`,
    encoded as UTF-8 and written without insignificant whitespace. The `files` of
    its `PackageAttributeInfo` include those of every patch, e.g. their `file`.

    Args:
      name: A unique name for this target.
      patches: Required. A list of `patch` targets. Must contain at least one
        entry.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    if not patches:
        fail("patches must contain at least one entry")

    _patch_set(
        # `_patch_set` attributes.
        name = name,
        patches = patches,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        # This should be package_metadata, but we use the legacy name
        # to support bazel 7.
        applicable_licenses = [],
    )
