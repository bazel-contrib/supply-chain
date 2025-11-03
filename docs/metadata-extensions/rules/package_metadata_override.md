<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `package_metadata_override`.

<a id="package_metadata_override"></a>

## package_metadata_override

<pre>
load("@package_metadata_extensions//rules:package_metadata_override.bzl", "package_metadata_override")

package_metadata_override(*, <a href="#package_metadata_override-name">name</a>, <a href="#package_metadata_override-metadata">metadata</a>, <a href="#package_metadata_override-packages">packages</a>, <a href="#package_metadata_override-visibility">visibility</a>)
</pre>



**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="package_metadata_override-name"></a>name |  A unique name for this macro instance. Normally, this is also the name for the macro's main or only target. The names of any other targets that this macro might create will be this name with a string suffix.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="package_metadata_override-metadata"></a>metadata |  A `package_metadata` target to use for the provided packages.<br><br>This overrides any `package_metadata` directly declared by the packages.   | <a href="https://bazel.build/concepts/labels">Label</a>; <a href="https://bazel.build/reference/be/common-definitions#configurable-attributes">nonconfigurable</a> | required |  |
| <a id="package_metadata_override-packages"></a>packages |  A list of packages the override applies to.<br><br>This follows the same syntax as [package_group](https://bazel.build/reference/be/functions#package_group), with the notable exception that `Label`s may refer to other repositories.   | List of strings; <a href="https://bazel.build/reference/be/common-definitions#configurable-attributes">nonconfigurable</a> | required |  |
| <a id="package_metadata_override-visibility"></a>visibility |  The visibility to be passed to this macro's exported targets. It always implicitly includes the location where this macro is instantiated, so this attribute only needs to be explicitly set if you want the macro's targets to be additionally visible somewhere else.   | <a href="https://bazel.build/concepts/labels">List of labels</a>; <a href="https://bazel.build/reference/be/common-definitions#configurable-attributes">nonconfigurable</a> | optional |  |


