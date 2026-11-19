<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `ProvenanceInfo`.

<a id="ProvenanceInfo"></a>

## ProvenanceInfo

<pre>
load("@package_metadata//providers:provenance_info.bzl", "ProvenanceInfo")

ProvenanceInfo(<a href="#ProvenanceInfo-provenance">provenance</a>, <a href="#ProvenanceInfo-files">files</a>)
</pre>

Provider for describing provenance info.

This is typically emitted by every target in the dependency graph of an artifact
and contains the `PackageMetadataInfo`s directly attached to the target as well
as `ProvenanceInfo` from dependencies of the target and their relationship to
the current `ProvenanceInfo`.

`ProvenanceInfo` provides information about a single node in the
(configured) target graph, including outgoing edges to its direct dependencies.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="ProvenanceInfo-provenance"></a>provenance | The [File](https://bazel.build/rules/lib/builtins/File) containing the provenance information. | none |
| <a id="ProvenanceInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the target, including transitive files from all dependencies. | `[]` |


