"""Utilities for creating target metadata."""

load("//providers:target_metadata_info.bzl", "TargetMetadataInfo")

visibility([
    "//common/...",
])

def create_target_metadata(*, actions, label, package_metadata = []):
    """Creates a TargetMetadataInfo provider with JSON metadata.

    This function generates a JSON file containing metadata about a Bazel target,
    including its label and references to package metadata from its dependencies.
    The metadata is structured for consumption by supply chain analysis tools.

    **Example:**

    ```starlark
    def _my_rule_impl(ctx):
        info = create_target_metadata(
            actions = ctx.actions,
            label = ctx.label,
            package_metadata = [
                dep[PackageMetadataInfo]
                for dep in ctx.attr.deps
                if PackageMetadataInfo in dep
            ],
        )
        return [info]
    ```

    Args:
        actions: The [actions](https://bazel.build/rules/lib/builtins/actions)
            object from the rule context, used to declare and write files.
        label: The [Label](https://bazel.build/rules/lib/builtins/Label) of the
            target being processed.
        package_metadata: A list of [PackageMetadataInfo](//providers:package_metadata_info.bzl)
            providers directly attached to the target being processed Defaults
            to an empty list.

    Returns:
        A [TargetMetadataInfo](//providers:target_metadata_info.bzl) provider
        containing the generated metadata file and transitive files from all
        package metadata.
    """
    metadata = actions.declare_file("{}.target-metadata.json".format(label.name))
    actions.write(
        output = metadata,
        content = json.encode({
            "label": str(label),
            "package_metadata": [p.metadata.path for p in package_metadata],
        }),
    )

    return TargetMetadataInfo(
        metadata = metadata,
        files = [a.files for a in package_metadata],
    )
