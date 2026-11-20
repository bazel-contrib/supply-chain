"""Declares rule `component_type`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `component_type` targets.
KIND = "build.bazel.attribute.component_type"

# The types supported by every SBOM format: the CycloneDX `component.type`
# values that also exist as SPDX 2.3 `primaryPackagePurpose` and SPDX 3.0.1
# `software_primaryPurpose` values.
TYPES = [
    "application",
    "container",
    "device",
    "file",
    "firmware",
    "framework",
    "library",
    "operating-system",
]

def _component_type_impl(ctx):
    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        output = output,
        content = json.encode(ctx.attr.type),
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

# Documentation lives on the public `component_type` macro below: Stardoc renders
# the macro, not this private rule.
_component_type = rule(
    implementation = _component_type_impl,
    attrs = {
        "type": attr.string(
            mandatory = True,
            values = TYPES,
        ),
    },
)

def component_type(
        # Disallow unnamed attributes.
        *,
        # `_component_type` attributes.
        name,
        type,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring the component type of a package or target.

    This is typically passed to the `attributes` of a `package_metadata` target.

    The type tells SBOM consumers what kind of thing the package is. Tools for
    building SBOMs report it as the CycloneDX `component.type`, the SPDX 2.3
    `primaryPackagePurpose` and the SPDX 3.0.1 `software_primaryPurpose`. Only the
    types supported by all of these formats are accepted. When a package does not
    declare one, CycloneDX reports `library` and the SPDX formats report no
    purpose.

    Usage:

    ```starlark
    load("@package_metadata//attributes:component_type.bzl", "component_type")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            ":component_type",
        ],
        purl = "pkg:oci/my-image?repository_url=registry.example.com%2Fmy-image",
        visibility = ["//visibility:public"],
    )

    component_type(
        name = "component_type",
        type = "container",
    )
    ```

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.component_type`, and its `attributes` file holds the
    `type` as a JSON string, written without insignificant whitespace:

    ```json
    "container"
    ```

    Args:
      name: A unique name for this target.
      type: Required. The type of the component, one of `application`,
        `container`, `device`, `file`, `firmware`, `framework`, `library` or
        `operating-system`.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    _component_type(
        # `_component_type` attributes.
        name = name,
        type = type,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        # This should be package_metadata, but we use the legacy name
        # to support bazel 7.
        applicable_licenses = [],
    )
