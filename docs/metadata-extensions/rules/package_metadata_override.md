<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `package_metadata_override`.

<a id="package_metadata_override"></a>

## package_metadata_override

<pre>
load("@package_metadata_extensions//rules:package_metadata_override.bzl", "package_metadata_override")

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


