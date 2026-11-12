<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `patch_set`.

<a id="patch_set"></a>

## patch_set

<pre>
load("@package_metadata//attributes:patch_set.bzl", "patch_set")

patch_set(*, <a href="#patch_set-name">name</a>, <a href="#patch_set-patches">patches</a>, <a href="#patch_set-tags">tags</a>, <a href="#patch_set-visibility">visibility</a>)
</pre>

Rule for declaring the set of patches applied to a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.
Each entry of `patches` is a `patch` target.

Usage:

```starlark
load("@package_metadata//attributes:patch.bzl", "patch")
load("@package_metadata//attributes:patch_set.bzl", "patch_set")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":patches",
    ],
    purl = "pkg:generic/ntp@4.2.8",
    visibility = ["//visibility:public"],
)

patch_set(
    name = "patches",
    patches = [":fix_overflow"],
)

patch(
    name = "fix_overflow",
    description = "Fixes a buffer overflow in the parser.",
    integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
    type = "backport",
    url = "https://example.com/patches/fix_overflow.patch",
)
```

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.patch_set`, and its `attributes` file holds a JSON
array of the objects described in `patch`, in the order of `patches`,
encoded as UTF-8 and written without insignificant whitespace. The `files` of
its `PackageAttributeInfo` include those of every patch, e.g. their `file`.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="patch_set-name"></a>name |  A unique name for this target.   |  none |
| <a id="patch_set-patches"></a>patches |  Required. A list of `patch` targets. Must contain at least one entry.   |  none |
| <a id="patch_set-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="patch_set-visibility"></a>visibility |  The visibility of this target.   |  `None` |


