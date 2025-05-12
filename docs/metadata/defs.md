<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of `@package_metadata`.

<a id="PackageAttributeInfo"></a>

## PackageAttributeInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageAttributeInfo")

PackageAttributeInfo(<a href="#PackageAttributeInfo-kind">kind</a>, <a href="#PackageAttributeInfo-attributes">attributes</a>, <a href="#PackageAttributeInfo-files">files</a>)
</pre>

Provider for declaring metadata about a Bazel package.

**Fields in this provider are not covered by the stability gurantee.**

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageAttributeInfo-kind"></a>kind | The identifier of the attribute.<br><br>This should generally be in reverse DNS format (e.g., `com.example.foo`). | none |
| <a id="PackageAttributeInfo-attributes"></a>attributes | The [File](https://bazel.build/rules/lib/builtins/File) containing the attributes.<br><br>The format of this file depends on the `kind` of attribute. Please consult the documentation of the attribute. | none |
| <a id="PackageAttributeInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s containing information about this attribute. | `[]` |


<a id="PackageMetadataInfo"></a>

## PackageMetadataInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageMetadataInfo")

PackageMetadataInfo(<a href="#PackageMetadataInfo-metadata">metadata</a>, <a href="#PackageMetadataInfo-files">files</a>)
</pre>

Provider for declaring metadata about a Bazel package.

**Fields in this provider are not covered by the stability gurantee.**

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataInfo-metadata"></a>metadata | The [File](https://bazel.build/rules/lib/builtins/File) containing metadata about the package. | none |
| <a id="PackageMetadataInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the package, including transitive files from all attributes of the package. | `[]` |


<a id="package_metadata"></a>

## package_metadata

<pre>
load("@package_metadata//:defs.bzl", "package_metadata")

package_metadata(*, <a href="#package_metadata-name">name</a>, <a href="#package_metadata-purl">purl</a>, <a href="#package_metadata-attributes">attributes</a>, <a href="#package_metadata-visibility">visibility</a>)
</pre>



**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="package_metadata-purl"></a>purl |  <p align="center"> - </p>   |  none |
| <a id="package_metadata-attributes"></a>attributes |  <p align="center"> - </p>   |  `[]` |
| <a id="package_metadata-visibility"></a>visibility |  <p align="center"> - </p>   |  `None` |


