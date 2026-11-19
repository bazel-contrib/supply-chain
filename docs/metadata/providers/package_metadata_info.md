<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `PackageMetadataInfo`.

<a id="PackageMetadataInfo"></a>

## PackageMetadataInfo

<pre>
load("@package_metadata//providers:package_metadata_info.bzl", "PackageMetadataInfo")

PackageMetadataInfo(<a href="#PackageMetadataInfo-metadata">metadata</a>, <a href="#PackageMetadataInfo-files">files</a>, <a href="#PackageMetadataInfo-attribute_kinds">attribute_kinds</a>)
</pre>

Provider for declaring metadata about a Bazel package.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataInfo-metadata"></a>metadata | The [File](https://bazel.build/rules/lib/builtins/File) containing metadata about the package. | none |
| <a id="PackageMetadataInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the package, including transitive files from all attributes of the package. | `[]` |
| <a id="PackageMetadataInfo-attribute_kinds"></a>attribute_kinds | A list of the `kind`s of the attributes declared by the package, so that tools can tell at analysis time which attributes a package has. The values of the attributes are only available in the `metadata` file. | `[]` |


