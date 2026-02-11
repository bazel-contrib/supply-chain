<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of `@package_metadata`.

<a id="PackageAttributeInfo"></a>

## PackageAttributeInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageAttributeInfo")

PackageAttributeInfo(<a href="#PackageAttributeInfo-kind">kind</a>, <a href="#PackageAttributeInfo-data">data</a>, <a href="#PackageAttributeInfo-files">files</a>)
</pre>

Provider for declaring metadata about a Bazel package.

> **Fields in this provider are not covered by the stability guarantee.**

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageAttributeInfo-kind"></a>kind | The identifier of the attribute.<br><br>This should generally be in reverse DNS format (e.g., `com.example.foo`). | none |
| <a id="PackageAttributeInfo-data"></a>data | The [File](https://bazel.build/rules/lib/builtins/File) containing the attribute specific data.<br><br>The format of this file depends on the `kind` of attribute. Please consult the documentation of the attribute. | none |
| <a id="PackageAttributeInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s containing information about this attribute. | `[]` |


<a id="PackageMetadataInfo"></a>

## PackageMetadataInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageMetadataInfo")

PackageMetadataInfo(<a href="#PackageMetadataInfo-metadata">metadata</a>, <a href="#PackageMetadataInfo-attributes">attributes</a>, <a href="#PackageMetadataInfo-files">files</a>)
</pre>

Provider for declaring metadata about a Bazel package.

> **Fields in this provider are not covered by the stability guarantee.**

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataInfo-metadata"></a>metadata | The [File](https://bazel.build/rules/lib/builtins/File) containing metadata about the package. | none |
| <a id="PackageMetadataInfo-attributes"></a>attributes | A [depset](https://bazel.build/rules/lib/builtins/depset) of `PackageAttributeInfo` providers. | `[]` |
| <a id="PackageMetadataInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the package, including transitive files from all attributes of the package. | `[]` |


<a id="PackageMetadataOverrideInfo"></a>

## PackageMetadataOverrideInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageMetadataOverrideInfo")

PackageMetadataOverrideInfo(*, <a href="#PackageMetadataOverrideInfo-packages">packages</a>, <a href="#PackageMetadataOverrideInfo-metadata">metadata</a>)
</pre>

Defines an override for `PackageMetadataInfo` for a set of packages.

> **Fields in this provider are not covered by the stability guarantee.**

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="PackageMetadataOverrideInfo-packages"></a>packages | A [PackageSpecificationInfo](https://bazel.build/rules/lib/providers/PackageSpecificationInfo) provider declaring which packages the override applies to.<br><br>This is typically created by a [package_group](https://bazel.build/rules/lib/globals/build#package_group) target. |
| <a id="PackageMetadataOverrideInfo-metadata"></a>metadata | The `PackageMetadataInfo` provider to use instead of the provider declared by package itself. |


<a id="PackageMetadataToolchainInfo"></a>

## PackageMetadataToolchainInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageMetadataToolchainInfo")

PackageMetadataToolchainInfo(<a href="#PackageMetadataToolchainInfo-metadata_overrides">metadata_overrides</a>)
</pre>

Toolchain for `package_metadata`.

> **Fields in this provider are not covered by the stability guarantee.**

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataToolchainInfo-metadata_overrides"></a>metadata_overrides | A sequence of `PackageMetadataOverrideInfo` providers. | `[]` |


<a id="package_metadata"></a>

## package_metadata

<pre>
load("@package_metadata//:defs.bzl", "package_metadata")

package_metadata(*, <a href="#package_metadata-name">name</a>, <a href="#package_metadata-purl">purl</a>, <a href="#package_metadata-attributes">attributes</a>, <a href="#package_metadata-visibility">visibility</a>, <a href="#package_metadata-tags">tags</a>)
</pre>



**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="package_metadata-purl"></a>purl |  <p align="center"> - </p>   |  none |
| <a id="package_metadata-attributes"></a>attributes |  <p align="center"> - </p>   |  `[]` |
| <a id="package_metadata-visibility"></a>visibility |  <p align="center"> - </p>   |  `None` |
| <a id="package_metadata-tags"></a>tags |  <p align="center"> - </p>   |  `None` |


<a id="purl.bazel"></a>

## purl.bazel

<pre>
load("@package_metadata//:defs.bzl", "purl")

purl.bazel(<a href="#purl.bazel-name">name</a>, <a href="#purl.bazel-version">version</a>)
</pre>

Defines a `purl` for a Bazel module.

This is typically used to construct `purl` for `package_metadata` targets in
Bazel modules.

This is **NOT** supported in `WORKSPACE` mode.

Example:

```starlark
load("@package_metadata//purl:purl.bzl", "purl")

package_metadata(
    name = "package_metadata",
    purl = purl.bazel(module_name(), module_version()),
    attributes = [
        # ...
    ],
    visibility = ["//visibility:public"],
)
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="purl.bazel-name"></a>name |  The name of the Bazel module. Typically [module_name()](https://bazel.build/rules/lib/globals/build#module_name).   |  none |
| <a id="purl.bazel-version"></a>version |  The version of the Bazel module. Typically [module_version()](https://bazel.build/rules/lib/globals/build#module_version). May be empty or `None`.   |  none |

**RETURNS**

The `purl` for the Bazel module (e.g. `pkg:bazel/foo` or
  `pkg:bazel/bar@1.2.3`).


<a id="purl.builder"></a>

## purl.builder

<pre>
load("@package_metadata//:defs.bzl", "purl")

purl.builder()
</pre>





