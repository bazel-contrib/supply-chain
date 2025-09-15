<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `PackageMetadataOverrideInfo`.

<a id="PackageMetadataOverrideInfo"></a>

## PackageMetadataOverrideInfo

<pre>
load("@package_metadata//providers:package_metadata_override_info.bzl", "PackageMetadataOverrideInfo")

PackageMetadataOverrideInfo(*, <a href="#PackageMetadataOverrideInfo-packages">packages</a>, <a href="#PackageMetadataOverrideInfo-metadata">metadata</a>)
</pre>

Defines an override for `PackageMetadataInfo` for a set of packages.

> **Fields in this provider are not covered by the stability gurantee.**

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="PackageMetadataOverrideInfo-packages"></a>packages | A [PackageSpecificationInfo](https://bazel.build/rules/lib/providers/PackageSpecificationInfo) provider declaring which packages the override applies to.<br><br>This is typically created by a [package_group](https://bazel.build/rules/lib/globals/build#package_group) target. |
| <a id="PackageMetadataOverrideInfo-metadata"></a>metadata | The `PackageMetadataInfo` provider to use instead of the provider declared by package itself. |


