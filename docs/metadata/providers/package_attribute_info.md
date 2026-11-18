<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `PackageAttributeInfo`.

<a id="PackageAttributeInfo"></a>

## PackageAttributeInfo

<pre>
load("@package_metadata//providers:package_attribute_info.bzl", "PackageAttributeInfo")

PackageAttributeInfo(<a href="#PackageAttributeInfo-kind">kind</a>, <a href="#PackageAttributeInfo-attributes">attributes</a>, <a href="#PackageAttributeInfo-files">files</a>, <a href="#PackageAttributeInfo-allow_multiple">allow_multiple</a>)
</pre>

Provider for declaring a single attribute of a Bazel package (e.g., the license
it is available under or the copyright notice).

Attributes are attached to a package by passing targets providing this to the
`attributes` of a `package_metadata` target. Declaring one is the extension
point for organizations that need to inject metadata beyond the attributes
provided by `@package_metadata`.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageAttributeInfo-kind"></a>kind | The identifier of the attribute.<br><br>This should generally be in reverse DNS format (e.g., `com.example.foo`). | none |
| <a id="PackageAttributeInfo-attributes"></a>attributes | The [File](https://bazel.build/rules/lib/builtins/File) containing the attributes.<br><br>The format of this file depends on the `kind` of attribute. Please consult the documentation of the attribute (e.g., `license` documents the JSON object it writes). | none |
| <a id="PackageAttributeInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s containing information about this attribute. | `[]` |
| <a id="PackageAttributeInfo-allow_multiple"></a>allow_multiple | Whether a `package_metadata` target may declare more than one attribute of this `kind`. Defaults to `False`.<br><br>If `False`, declaring two attributes of this `kind` on the same `package_metadata` fails the build. If `True`, all of them are recorded, in the order they were declared, and consumers are expected to handle a list of `attributes` files for this `kind` (e.g., a `patch` attribute declared once per patch). | `False` |


