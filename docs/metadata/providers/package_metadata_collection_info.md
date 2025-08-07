<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `PackageMetadataCollectionInfo`.

<a id="PackageMetadataCollectionInfo"></a>

## PackageMetadataCollectionInfo

<pre>
load("@package_metadata//providers:package_metadata_collection_info.bzl", "PackageMetadataCollectionInfo")

PackageMetadataCollectionInfo(<a href="#PackageMetadataCollectionInfo-_init-dependencies">dependencies</a>, <a href="#PackageMetadataCollectionInfo-_init-tools">tools</a>, <a href="#PackageMetadataCollectionInfo-_init-deps">deps</a>)
</pre>

Provides a collection of `PackageMetadataInfo`s transitively used by target.

> **Fields in this provider are not covered by the stability gurantee.**

**CONSTRUCTOR PARAMETERS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataCollectionInfo-_init-dependencies"></a>dependencies | A [depset](https://bazel.build/rules/lib/builtins/depset) of `PackageMetadataInfo`s that are dependencies of the targets.<br><br>This only includes dependencies that are used at runtime (i.e., from `cfg = "target"`). For dependencies that are needed at compile time (e.g., compilers, linters, ...), see `tools`. | `[]` |
| <a id="PackageMetadataCollectionInfo-_init-tools"></a>tools | A [depset](https://bazel.build/rules/lib/builtins/depset) of `PackageMetadataInfo`s that are are needed to build the target.<br><br>For dependencies that are needed at runtime time see `dependencies`. | `[]` |
| <a id="PackageMetadataCollectionInfo-_init-deps"></a>deps | <p align="center">-</p> | `[]` |

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="PackageMetadataCollectionInfo-dependencies"></a>dependencies |  A [depset](https://bazel.build/rules/lib/builtins/depset) of `PackageMetadataInfo`s that are dependencies of the targets.<br><br>This only includes dependencies that are used at runtime (i.e., from `cfg = "target"`). For dependencies that are needed at compile time (e.g., compilers, linters, ...), see `tools`.    |
| <a id="PackageMetadataCollectionInfo-files"></a>files |  A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s referenced by `PackageMetadataInfo` providers in this collection.    |
| <a id="PackageMetadataCollectionInfo-tools"></a>tools |  A [depset](https://bazel.build/rules/lib/builtins/depset) of `PackageMetadataInfo`s that are are needed to build the target.<br><br>For dependencies that are needed at runtime time see `dependencies`.    |


