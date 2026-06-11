<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Common utilities for creating package and target metadata.

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

<a id="package_metadata_common.create_package_metadata"></a>

## package_metadata_common.create_package_metadata

<pre>
load("@package_metadata//common:common.bzl", "package_metadata_common")

package_metadata_common.create_package_metadata(*, <a href="#package_metadata_common.create_package_metadata-actions">actions</a>, <a href="#package_metadata_common.create_package_metadata-label">label</a>, <a href="#package_metadata_common.create_package_metadata-purl">purl</a>, <a href="#package_metadata_common.create_package_metadata-attributes">attributes</a>)
</pre>

Creates a PackageMetadataInfo provider with JSON metadata.

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


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata_common.create_package_metadata-actions"></a>actions |  The [actions](https://bazel.build/rules/lib/builtins/actions) object from the rule context, used to declare and write files.   |  none |
| <a id="package_metadata_common.create_package_metadata-label"></a>label |  The [Label](https://bazel.build/rules/lib/builtins/Label) of the target being processed.   |  none |
| <a id="package_metadata_common.create_package_metadata-purl"></a>purl |  A string containing the [PURL](https://github.com/package-url/purl-spec) uniquely identifying this package (e.g., "pkg:npm/lodash@4.17.21").   |  none |
| <a id="package_metadata_common.create_package_metadata-attributes"></a>attributes |  A list of [PackageAttributeInfo](//providers:package_attribute_info.bzl) providers representing package attributes (e.g., source location, license). Defaults to an empty list.   |  `[]` |

**RETURNS**

A [PackageMetadataInfo](//providers:package_metadata_info.bzl) provider
  containing the generated metadata file and transitive files from all
  attributes.


<a id="package_metadata_common.create_target_metadata"></a>

## package_metadata_common.create_target_metadata

<pre>
load("@package_metadata//common:common.bzl", "package_metadata_common")

package_metadata_common.create_target_metadata(*, <a href="#package_metadata_common.create_target_metadata-actions">actions</a>, <a href="#package_metadata_common.create_target_metadata-label">label</a>, <a href="#package_metadata_common.create_target_metadata-package_metadata">package_metadata</a>)
</pre>

Creates a TargetMetadataInfo provider with JSON metadata.

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


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata_common.create_target_metadata-actions"></a>actions |  The [actions](https://bazel.build/rules/lib/builtins/actions) object from the rule context, used to declare and write files.   |  none |
| <a id="package_metadata_common.create_target_metadata-label"></a>label |  The [Label](https://bazel.build/rules/lib/builtins/Label) of the target being processed.   |  none |
| <a id="package_metadata_common.create_target_metadata-package_metadata"></a>package_metadata |  A list of [PackageMetadataInfo](//providers:package_metadata_info.bzl) providers directly attached to the target being processed Defaults to an empty list.   |  `[]` |

**RETURNS**

A [TargetMetadataInfo](//providers:target_metadata_info.bzl) provider
  containing the generated metadata file and transitive files from all
  package metadata.


