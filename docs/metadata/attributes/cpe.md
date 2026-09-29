<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `cpe`.

<a id="cpe"></a>

## cpe

<pre>
load("@package_metadata//attributes:cpe.bzl", "cpe")

cpe(*, <a href="#cpe-name">name</a>, <a href="#cpe-identifiers">identifiers</a>, <a href="#cpe-tags">tags</a>, <a href="#cpe-visibility">visibility</a>)
</pre>

Rule for declaring the CPEs of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

`identifiers` entries are assumed to be UTF-8 encoded. Tools for building SBOMs
presume that.

Usage:

```starlark
load("@package_metadata//attributes:cpe.bzl", "cpe")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":cpe",
    ],
    purl = "pkg:generic/ntp@4.2.8",
    visibility = ["//visibility:public"],
)

cpe(
    name = "cpe",
    identifiers = [
        "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*",
    ],
)
```

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.cpe`, and its `attributes` file holds the `identifiers`
entries as a JSON array of strings, encoded as UTF-8 and written without
insignificant whitespace:

```json
[
    "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*"
]
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="cpe-name"></a>name |  A unique name for this target.   |  none |
| <a id="cpe-identifiers"></a>identifiers |  Required. One or more [CPE](https://en.wikipedia.org/wiki/Common_Platform_Enumeration) identifiers of the package. UTF-8 encoding.   |  none |
| <a id="cpe-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="cpe-visibility"></a>visibility |  The visibility of this target.   |  `None` |


