"""Declares rule `patch`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `patch` targets.
KIND = "build.bazel.attribute.patch"

# Mirrors the CycloneDX `pedigree.patches[].type` values. SPDX 3.0.1 has no
# equivalent, tools may drop it.
PATCH_TYPES = ["unofficial", "monkey", "backport", "cherry-pick"]

PatchInfo = provider(
    doc = "Provider carrying the fields of a single `patch` target, consumed by `patch_set`.",
    fields = {
        "description": "The description of the patch.",
        "file": "The short path of the patch file, or None.",
        "integrity": "The list of integrity strings of the patch.",
        "resolves": "The list of issues resolved by the patch (JSON-serializable), or empty.",
        "type": "The type of the patch.",
        "url": "The URL of the patch, or None.",
    },
)

def _patch_impl(ctx):
    if not ctx.attr.integrity:
        fail("integrity must contain at least one entry")
    if bool(ctx.attr.url) == bool(ctx.file.file):
        fail("exactly one of url or file must be set")

    value = {
        "description": ctx.attr.description,
        "integrity": ctx.attr.integrity,
        "type": ctx.attr.type,
    }
    if ctx.attr.url:
        value["url"] = ctx.attr.url
    if ctx.file.file:
        value["file"] = ctx.file.file.short_path
    if ctx.attr.resolves:
        value["resolves"] = ctx.attr.resolves

    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        output = output,
        content = json.encode(value),
    )
    return [
        DefaultInfo(
            files = depset(direct = [output]),
        ),
        PackageAttributeInfo(
            kind = KIND,
            attributes = output,
            files = [depset(direct = [ctx.file.file])] if ctx.file.file else [],
        ),
        PatchInfo(
            description = ctx.attr.description,
            file = ctx.file.file.short_path if ctx.file.file else None,
            integrity = ctx.attr.integrity,
            resolves = ctx.attr.resolves,
            type = ctx.attr.type,
            url = ctx.attr.url or None,
        ),
    ]

# Documentation lives on the public `patch` macro below: Stardoc renders the
# macro, not this private rule.
_patch = rule(
    implementation = _patch_impl,
    attrs = {
        "description": attr.string(
            mandatory = True,
        ),
        "integrity": attr.string_list(
            mandatory = True,
        ),
        "file": attr.label(
            allow_single_file = True,
        ),
        "resolves": attr.string_list(),
        "type": attr.string(
            mandatory = True,
            values = PATCH_TYPES,
        ),
        "url": attr.string(),
    },
)

def patch(
        # Disallow unnamed attributes.
        *,
        # `_patch` attributes.
        name,
        description,
        integrity,
        type,
        url = None,
        file = None,
        resolves = None,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring a single patch applied to a package.

    Bazel attributes cannot be structs, so a patch is declared as its own
    target and grouped with `patch_set`. This target can also be passed on its
    own to the `attributes` of a `package_metadata` target.

    Values are passed through as-is: tools consuming the
    attribute are responsible for mapping them to the format they produce.
    Entries are assumed to be UTF-8 encoded.

    Usage:

    ```starlark
    load("@package_metadata//attributes:patch.bzl", "patch")

    patch(
        name = "fix_overflow",
        description = "Fixes a buffer overflow in the parser.",
        integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
        resolves = ["CVE-2024-12345"],
        type = "backport",
        url = "https://example.com/patches/fix_overflow.patch",
    )

    patch(
        name = "local_tweak",
        description = "Local tweak.",
        file = "patches/local_tweak.patch",
        integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
        type = "unofficial",
    )
    ```

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.patch`, and its `attributes` file holds a JSON object
    written without insignificant whitespace, with keys sorted. `url`, `file` and
    `resolves` are omitted when unset:

    ```json
    {
        "description": "Fixes a buffer overflow in the parser.",
        "integrity": ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
        "resolves": ["CVE-2024-12345"],
        "type": "backport",
        "url": "https://example.com/patches/fix_overflow.patch"
    }
    ```

    Args:
      name: A unique name for this target.
      description: Required. Human-readable description of what the patch does.
      integrity: Required. Checksums of the patch file as SRI strings
        (`<algorithm>-<base64 digest>`), e.g. `["sha256-<base64 digest>"]`.
        Must contain at least one entry.
      type: Required. The nature of the patch, one of `unofficial`, `monkey`,
        `backport` or `cherry-pick` (CycloneDX `pedigree.patches[].type`).
      url: Where the patch can be retrieved from. Exactly one of `url` or
        `file` must be set.
      file: A patch file in the workspace. Exactly one of `url` or `file` must
        be set. The JSON holds the file's `short_path`, and the file is added
        to the `files` of the `PackageAttributeInfo`.
      resolves: Optional. Identifiers (e.g., CVE ids, issue URLs) of the
        issues the patch resolves.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    if not integrity:
        fail("integrity must contain at least one entry")

    if bool(url) == bool(file):
        fail("exactly one of url or file must be set")

    _patch(
        # `_patch` attributes.
        name = name,
        description = description,
        integrity = integrity,
        type = type,
        url = url,
        file = file,
        resolves = resolves,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        # This should be package_metadata, but we use the legacy name
        # to support bazel 7.
        applicable_licenses = [],
    )
