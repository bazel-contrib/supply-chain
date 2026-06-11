"""Common utilities for creating package and target metadata.

This module provides helper functions for creating metadata about Bazel packages
and targets in a structured format. These utilities are typically used within
rule implementations to generate metadata providers.

## Usage

Load this module to access helper functions for creating metadata:

```starlark
load("@rules_supplychain//metadata/common:common.bzl", "package_metadata_common")

def _my_rule_impl(ctx):
    # Create package metadata
    pkg_info = package_metadata_common.create_package_metadata(
        actions = ctx.actions,
        label = ctx.label,
        purl = "pkg:example/my-package@1.0.0",
        attributes = [attr[PackageAttributeInfo] for attr in ctx.attr.attributes],
    )

    # Create target metadata
    target_info = package_metadata_common.create_target_metadata(
        actions = ctx.actions,
        label = ctx.label,
        package_metadata = [dep[PackageMetadataInfo] for dep in ctx.attr.deps],
    )

    return [pkg_info, target_info]
```
"""

load("//common/private:create_package_metadata.bzl", "create_package_metadata")
load("//common/private:create_target_metadata.bzl", "create_target_metadata")

visibility("public")

package_metadata_common = struct(
    create_package_metadata = create_package_metadata,
    create_target_metadata = create_target_metadata,
)
