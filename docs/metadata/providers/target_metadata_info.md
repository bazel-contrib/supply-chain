<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `TargetMetadataInfo`.

<a id="TargetMetadataInfo"></a>

## TargetMetadataInfo

<pre>
load("@package_metadata//providers:target_metadata_info.bzl", "TargetMetadataInfo")

TargetMetadataInfo(<a href="#TargetMetadataInfo-metadata">metadata</a>, <a href="#TargetMetadataInfo-files">files</a>)
</pre>

Provider for declaring metadata about a Bazel target.

> **Fields in this provider are not covered by the stability guarantee.**

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="TargetMetadataInfo-metadata"></a>metadata | The [File](https://bazel.build/rules/lib/builtins/File) containing metadata about the target. | none |
| <a id="TargetMetadataInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the target, including transitive files from all dependencies. | `[]` |


