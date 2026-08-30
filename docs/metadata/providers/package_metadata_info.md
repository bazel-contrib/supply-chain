<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `PackageMetadataInfo`.

<a id="PackageMetadataInfo"></a>

## PackageMetadataInfo

<pre>
load("@package_metadata//providers:package_metadata_info.bzl", "PackageMetadataInfo")

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


