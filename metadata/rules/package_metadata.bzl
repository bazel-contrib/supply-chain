"""Declares rule `package_metadata`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")
load("//providers:package_metadata_info.bzl", "PackageMetadataInfo")

visibility("public")

def _package_metadata_impl(ctx):
    attributes = [a[PackageAttributeInfo] for a in ctx.attr.attributes]

    metadata = ctx.actions.declare_file("{}.package-metadata.json".format(ctx.attr.name))

    ctx.actions.write(
        output = metadata,
        content = json.encode({
            "attributes": {a.kind: a.attributes.path for a in attributes},
            "label": str(ctx.label),
            "purl": ctx.attr.purl,
        }),
    )

    return [
        DefaultInfo(
            files = depset(
                direct = [
                    metadata,
                ],
            ),
        ),
        PackageMetadataInfo(
            metadata = metadata,
            files = [a.files for a in attributes],
        ),
    ]

# Documentation lives on the public `package_metadata` macro below: Stardoc
# renders the macro, not this private rule.
_package_metadata = rule(
    implementation = _package_metadata_impl,
    attrs = {
        "attributes": attr.label_list(
            mandatory = False,
            providers = [
                PackageAttributeInfo,
            ],
        ),
        "purl": attr.string(
            mandatory = True,
        ),
    },
    provides = [
        PackageMetadataInfo,
    ],
)

def package_metadata(
        # Disallow unnamed attributes.
        *,
        # `_package_metadata` attributes.
        name,
        purl,
        attributes = [],
        # Common attributes (subset since this target is non-configurable).
        visibility = None,
        tags = None):
    """Rule for declaring `PackageMetadataInfo`, typically of a `bzlmod` module.

    A `package_metadata` target identifies a package and where it was retrieved
    from. Additional metadata (e.g., the license the package is under) is
    attached by passing `attributes`.

    Targets are associated with a `package_metadata` target module-wide via
    `repo(default_package_metadata = ...)` in `REPO.bazel`, package-wide via
    `package(default_package_metadata = ...)`, or individually via the
    `package_metadata` attribute of the target.

    Usage:

    ```starlark
    load("@package_metadata//purl:purl.bzl", "purl")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            ":license",
        ],
        purl = purl.bazel(module_name(), module_version()),
        visibility = ["//visibility:public"],
    )
    ```

    Format: the `metadata` file of the `PackageMetadataInfo` of this rule holds
    a JSON object, encoded as UTF-8 and written without insignificant whitespace
    (the example below is indented for readability):

    ```json
    {
        "attributes": {
            "build.bazel.attribute.license": "bazel-out/k8-fastbuild/bin/license.package-attribute.json"
        },
        "label": "@@//:package_metadata",
        "purl": "pkg:bazel/package_metadata"
    }
    ```

    | Field | Description |
    | :---- | :---------- |
    | `attributes` | The declared `attributes`, keyed by the `kind` of their `PackageAttributeInfo`. Values are the paths of the `attributes` files, relative to the execution root. |
    | `label` | The label of this `package_metadata` target. |
    | `purl` | The `purl` of this target, verbatim. |

    Since `attributes` is keyed by `kind`, declaring two attributes of the same
    `kind` on one target is not meaningful — only one of them is recorded.

    Args:
      name: A unique name for this target.
      purl: Required. The [PURL](https://github.com/package-url/purl-spec)
        uniquely identifying this package.

        For Bazel modules, this is typically constructed with `purl.bazel`.
      attributes: A list of `attributes` of the package (e.g., source location,
        license, ...).

        Each element must be a target providing `PackageAttributeInfo` (e.g., a
        `license` target).
      visibility: The visibility of this target.

        `package_metadata` targets are typically `//visibility:public` so that
        consumers of the module can read the metadata.
      tags: A list of arbitrary tags to apply to this target.
    """

    _package_metadata(
        # `_package_metadata` attributes.
        name = name,
        purl = purl,
        attributes = attributes,

        # Common attributes.
        visibility = visibility,
        tags = tags,
        applicable_licenses = [],
    )
