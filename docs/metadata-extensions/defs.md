<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of `@package_metadata_extensions`.

<a id="package_metadata_override"></a>

## package_metadata_override

<pre>
load("@package_metadata_extensions//:defs.bzl", "package_metadata_override")

package_metadata_override(*, <a href="#package_metadata_override-name">name</a>, <a href="#package_metadata_override-metadata">metadata</a>, <a href="#package_metadata_override-targets">targets</a>, <a href="#package_metadata_override-visibility">visibility</a>)
</pre>



**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="package_metadata_override-name"></a>name |  A unique name for this macro instance. Normally, this is also the name for the macro's main or only target. The names of any other targets that this macro might create will be this name with a string suffix.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="package_metadata_override-metadata"></a>metadata |  A `package_metadata` target to use for the provided packages.<br><br>This overrides any `package_metadata` directly declared by the packages.   | <a href="https://bazel.build/concepts/labels">Label</a>; <a href="https://bazel.build/reference/be/common-definitions#configurable-attributes">nonconfigurable</a> | required |  |
| <a id="package_metadata_override-targets"></a>targets |  A list of targets the override applies to.<br><br>This allows passing targets of the form<br><br>  - `//foo/...` matches all targets under `//foo` in the current repository.   - `//foo` matches all targets in `//foo` (but not in subpackages).   - `//foo:bar` matches exactly target `//foo:bar`.<br><br>This also supports referencing packages and targets in external repositories.<br><br>  - `@other_repo//foo/...` matches all targets under `//foo` in `other_repo`.   - `@other_repo//foo` matches all targets in `//foo` in `other_repo` (but not     in subpackages).   - `@other_repo//foo:bar` matches exactly target `//foo:bar` in `other_repo`.   | List of strings; <a href="https://bazel.build/reference/be/common-definitions#configurable-attributes">nonconfigurable</a> | required |  |
| <a id="package_metadata_override-visibility"></a>visibility |  The visibility to be passed to this macro's exported targets. It always implicitly includes the location where this macro is instantiated, so this attribute only needs to be explicitly set if you want the macro's targets to be additionally visible somewhere else.   | <a href="https://bazel.build/concepts/labels">List of labels</a>; <a href="https://bazel.build/reference/be/common-definitions#configurable-attributes">nonconfigurable</a> | optional |  |


