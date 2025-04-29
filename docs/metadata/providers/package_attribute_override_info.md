<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `PackageAttributeOverrideInfo`.

<a id="PackageAttributeOverrideInfo"></a>

## PackageAttributeOverrideInfo

<pre>
load("@package_metadata//providers:package_attribute_override_info.bzl", "PackageAttributeOverrideInfo")

PackageAttributeOverrideInfo(<a href="#PackageAttributeOverrideInfo-package">package</a>, <a href="#PackageAttributeOverrideInfo-overrides">overrides</a>)
</pre>

Provider for declaring overrides for attributes of `package_metadata` targets.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageAttributeOverrideInfo-package"></a>package | The [Label](https://bazel.build/rules/lib/builtins/Label) of the `package_metadata` target to override attributes of. | none |
| <a id="PackageAttributeOverrideInfo-overrides"></a>overrides | <p align="center"> - </p> | `[]` |


