"""Utilities for creating package metadata."""

load("//providers:package_metadata_info.bzl", "PackageMetadataInfo")

visibility([
    "//common/...",
])

def create_package_metadata(*, actions, label, purl, attributes = []):
    """Creates a PackageMetadataInfo provider with JSON metadata.

    This function generates a JSON file containing metadata about a Bazel package,
    including its PURL (Package URL), label, and attributes. The metadata is
    structured for consumption by supply chain analysis tools.

    **Example:**

    ```starlark
    def _my_rule_impl(ctx):
        info = create_package_metadata(
            actions = ctx.actions,
            label = ctx.label,
            purl = "pkg:npm/my-package@1.0.0",
            attributes = [a[PackageAttributeInfo] for a in ctx.attr.attributes],
        )
        return [info]
    ```

    Args:
        actions: The [actions](https://bazel.build/rules/lib/builtins/actions)
            object from the rule context, used to declare and write files.
        label: The [Label](https://bazel.build/rules/lib/builtins/Label) of the
            target being processed.
        purl: A string containing the [PURL](https://github.com/package-url/purl-spec)
            uniquely identifying this package (e.g., "pkg:npm/lodash@4.17.21").
        attributes: A list of [PackageAttributeInfo](//providers:package_attribute_info.bzl)
            providers representing package attributes (e.g., source location, license).
            Defaults to an empty list.

    Returns:
        A [PackageMetadataInfo](//providers:package_metadata_info.bzl) provider
        containing the generated metadata file and transitive files from all
        attributes.
    """

    metadata = actions.declare_file("{}.package-metadata.json".format(label.name))
    actions.write(
        output = metadata,
        content = json.encode({
            "attributes": {a.kind: a.attributes.path for a in attributes},
            "label": str(label),
            "purl": purl,
        }),
    )

    return PackageMetadataInfo(
        metadata = metadata,
        files = [a.files for a in attributes],
    )
