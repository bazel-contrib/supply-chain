<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Common utilities for creating package metadata.

These helpers are building blocks for rule authors who need to declare
`PackageMetadataInfo` from their own rules, rather than via the
`package_metadata` rule.

<a id="package_metadata_common.create_package_metadata"></a>

## package_metadata_common.create_package_metadata

<pre>
load("@package_metadata//common:common.bzl", "package_metadata_common")

package_metadata_common.create_package_metadata(*, <a href="#package_metadata_common.create_package_metadata-actions">actions</a>, <a href="#package_metadata_common.create_package_metadata-label">label</a>, <a href="#package_metadata_common.create_package_metadata-purl">purl</a>, <a href="#package_metadata_common.create_package_metadata-attributes">attributes</a>)
</pre>

Creates a `PackageMetadataInfo` provider with JSON metadata.

This function generates a JSON file containing metadata about a Bazel package,
including its PURL (Package URL), label, and attributes. The metadata is
structured for consumption by supply chain analysis tools.

Example:

```starlark
load("@package_metadata//common:common.bzl", "package_metadata_common")

def _my_rule_impl(ctx):
    info = package_metadata_common.create_package_metadata(
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
| <a id="package_metadata_common.create_package_metadata-attributes"></a>attributes |  A list of `PackageAttributeInfo` providers representing package attributes (e.g., source location, license). Defaults to an empty list.   |  `[]` |

**RETURNS**

A `PackageMetadataInfo` provider containing the generated metadata file and
  transitive files from all attributes.


