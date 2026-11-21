<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of `@package_metadata_extensions`.

<a id="PackageLinkingInfo"></a>

## PackageLinkingInfo

<pre>
load("@package_metadata_extensions//:defs.bzl", "PackageLinkingInfo")

PackageLinkingInfo(<a href="#PackageLinkingInfo-static_linked">static_linked</a>, <a href="#PackageLinkingInfo-dynamic_linked">dynamic_linked</a>)
</pre>

Classifies direct dependency targets by how the providing target links them.

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="PackageLinkingInfo-static_linked"></a>static_linked |  List of Labels for direct dependency targets this target links statically or embeds. These must be labels of direct dependencies, not transitive targets.    |
| <a id="PackageLinkingInfo-dynamic_linked"></a>dynamic_linked |  List of Labels for direct dependency targets this target links dynamically. These must be labels of direct dependencies, not transitive targets.    |


<a id="package_metadata_override"></a>

## package_metadata_override

<pre>
load("@package_metadata_extensions//:defs.bzl", "package_metadata_override")

package_metadata_override(<a href="#package_metadata_override-name">name</a>, <a href="#package_metadata_override-targets">targets</a>, <a href="#package_metadata_override-metadata">metadata</a>, <a href="#package_metadata_override-visibility">visibility</a>)
</pre>

Overrides package metadata for the matching target patterns.

**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata_override-name"></a>name |  Name for the generated override target.   |  none |
| <a id="package_metadata_override-targets"></a>targets |  A list of strings containing package or target patterns. For example, `//foo/...` matches descendants, `//foo` matches one package, and `//foo:bar` matches one target. Prefix a pattern with `-` to exclude it. Patterns may name external repositories, such as `@other_repo//foo/...`.   |  none |
| <a id="package_metadata_override-metadata"></a>metadata |  A target providing PackageMetadataInfo.   |  none |
| <a id="package_metadata_override-visibility"></a>visibility |  Visibility applied to generated targets.   |  `None` |


