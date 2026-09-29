<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `TargetInfo`.

<a id="TargetInfo"></a>

## TargetInfo

<pre>
load("@package_metadata//providers:target_info.bzl", "TargetInfo")

TargetInfo(<a href="#TargetInfo-_init-metadata">metadata</a>, <a href="#TargetInfo-_init-files">files</a>)
</pre>

Provider for describing a single target.

This includes the `PackageMetadataInfo`s directly attached to the target as well
as `TargetInfo` from dependencies of the target.

`TargetInfo` provides information about a single node in the
(configured) target graph, including outgoing edges to its direct dependencies.

**CONSTRUCTOR PARAMETERS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="TargetInfo-_init-metadata"></a>metadata | <p align="center">-</p> | none |
| <a id="TargetInfo-_init-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the target, including transitive files from all dependencies. | `[]` |

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="TargetInfo-files"></a>files |  A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the target, including transitive files from all dependencies.    |
| <a id="TargetInfo-info"></a>info |  The [File](https://bazel.build/rules/lib/builtins/File) containing the information about the target.    |


