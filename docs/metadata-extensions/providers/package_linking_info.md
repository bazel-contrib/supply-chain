<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Provider for per-target direct dependency linkage classifications.

Return this from a rule when the way it links a dependency depends on how the
target is configured, for example a flag choosing a static or a dynamic library.
The SBOM tooling then reports each dependency as `static_link` or `dynamic_link`.

<a id="PackageLinkingInfo"></a>

## PackageLinkingInfo

<pre>
load("@package_metadata_extensions//providers:package_linking_info.bzl", "PackageLinkingInfo")

PackageLinkingInfo(<a href="#PackageLinkingInfo-static_linked">static_linked</a>, <a href="#PackageLinkingInfo-dynamic_linked">dynamic_linked</a>)
</pre>

Classifies direct dependency targets by how the providing target links them.

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="PackageLinkingInfo-static_linked"></a>static_linked |  List of Labels for direct dependency targets this target links statically or embeds. These must be labels of direct dependencies, not transitive targets.    |
| <a id="PackageLinkingInfo-dynamic_linked"></a>dynamic_linked |  List of Labels for direct dependency targets this target links dynamically. These must be labels of direct dependencies, not transitive targets.    |


