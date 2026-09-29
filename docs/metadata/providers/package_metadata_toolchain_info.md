<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `PackageMetadataToolchainInfo`.

<a id="PackageMetadataToolchainInfo"></a>

## PackageMetadataToolchainInfo

<pre>
load("@package_metadata//providers:package_metadata_toolchain_info.bzl", "PackageMetadataToolchainInfo")

PackageMetadataToolchainInfo(<a href="#PackageMetadataToolchainInfo-metadata_overrides">metadata_overrides</a>)
</pre>

Toolchain for `package_metadata`.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataToolchainInfo-metadata_overrides"></a>metadata_overrides | A sequence of `PackageMetadataOverrideInfo` providers. | `[]` |


