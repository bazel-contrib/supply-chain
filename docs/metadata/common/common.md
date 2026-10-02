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


<a id="package_metadata_common.get_package_metadata"></a>

## package_metadata_common.get_package_metadata

<pre>
load("@package_metadata//common:common.bzl", "package_metadata_common")

package_metadata_common.get_package_metadata(*, <a href="#package_metadata_common.get_package_metadata-toolchain">toolchain</a>, <a href="#package_metadata_common.get_package_metadata-label">label</a>, <a href="#package_metadata_common.get_package_metadata-package_metadata">package_metadata</a>)
</pre>

Returns the `PackageMetadataInfo` providers that apply to a target.

Metadata declared by a target (usually via its `package_metadata` attribute)
can be overridden by the consumer of a module, e.g., to attach metadata to a
dependency that does not declare any itself. Such overrides are registered as
`PackageMetadataOverrideInfo` providers on the `package_metadata` toolchain,
each of them declaring the set of packages it applies to.

This function resolves both sources into the metadata to use for `label`: if
any override on `toolchain` matches `label`, the metadata of all matching
overrides is returned and the metadata declared by the target itself is
ignored. Otherwise, `package_metadata` is returned unchanged.

Example:

```starlark
load("@package_metadata//common:common.bzl", "package_metadata_common")
load("@package_metadata//providers:package_metadata_info.bzl", "PackageMetadataInfo")

_TOOLCHAIN_TYPE = "@package_metadata//toolchains:toolchain_type"

def _my_rule_impl(ctx):
    metadata = package_metadata_common.get_package_metadata(
        toolchain = ctx.toolchains[_TOOLCHAIN_TYPE],
        label = ctx.label,
        package_metadata = [m[PackageMetadataInfo] for m in ctx.attr.package_metadata],
    )
    # ...

my_rule = rule(
    implementation = _my_rule_impl,
    attrs = {
        "package_metadata": attr.label_list(
            providers = [PackageMetadataInfo],
        ),
    },
    toolchains = [
        config_common.toolchain_type(_TOOLCHAIN_TYPE, mandatory = False),
    ],
)
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata_common.get_package_metadata-toolchain"></a>toolchain |  The resolved `package_metadata` [toolchain](https://bazel.build/extending/toolchains), whose `package_metadata` field is a `PackageMetadataToolchainInfo` provider declaring the registered overrides. May be `None` if the toolchain is optional and was not resolved, in which case no overrides are applied.   |  none |
| <a id="package_metadata_common.get_package_metadata-label"></a>label |  The [Label](https://bazel.build/rules/lib/builtins/Label) of the target to look up metadata for. It is matched against the packages each override applies to.   |  none |
| <a id="package_metadata_common.get_package_metadata-package_metadata"></a>package_metadata |  A list of `PackageMetadataInfo` providers declared by the target itself, used if no override applies to `label`.   |  none |

**RETURNS**

A list of `PackageMetadataInfo` providers: the metadata of all overrides
  matching `label`, or `package_metadata` if there is no such override.


