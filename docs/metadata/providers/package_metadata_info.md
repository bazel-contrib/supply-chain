<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `PackageMetadataInfo`.

<a id="PackageMetadataInfo"></a>

## PackageMetadataInfo

<pre>
load("@package_metadata//providers:package_metadata_info.bzl", "PackageMetadataInfo")

PackageMetadataInfo(<a href="#PackageMetadataInfo-_init-metadata">metadata</a>, <a href="#PackageMetadataInfo-_init-purl">purl</a>, <a href="#PackageMetadataInfo-_init-files">files</a>)
</pre>

Provider for declaring metadata about a Bazel package.

> **Fields in this provider are not covered by the stability gurantee.**

**CONSTRUCTOR PARAMETERS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataInfo-_init-metadata"></a>metadata | The [File](https://bazel.build/rules/lib/builtins/File) containing metadata about the package. | none |
| <a id="PackageMetadataInfo-_init-purl"></a>purl | PURL | none |
| <a id="PackageMetadataInfo-_init-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the package, including transitive files from all attributes of the package. | `[]` |

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="PackageMetadataInfo-kind"></a>kind |  Type descriminator    |
| <a id="PackageMetadataInfo-files"></a>files |  A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the package, including transitive files from all attributes of the package.    |
| <a id="PackageMetadataInfo-metadata"></a>metadata |  The [File](https://bazel.build/rules/lib/builtins/File) containing metadata about the package.    |
| <a id="PackageMetadataInfo-purl"></a>purl |  PURL    |


