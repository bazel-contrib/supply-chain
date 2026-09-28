<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `copyright_notice`.

<a id="copyright_notice"></a>

## copyright_notice

<pre>
load("@package_metadata//attributes:copyright_notice.bzl", "copyright_notice")

copyright_notice(*, <a href="#copyright_notice-name">name</a>, <a href="#copyright_notice-text">text</a>, <a href="#copyright_notice-tags">tags</a>, <a href="#copyright_notice-visibility">visibility</a>)
</pre>

Rule for declaring the copyright notice of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

Copyright `text` entries are assumed to be UTF-8 encoded. Tools for building
SBOMs presume that.

Usage:

```starlark
load("@package_metadata//attributes:copyright_notice.bzl", "copyright_notice")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":copyright_notice",
    ],
    purl = "pkg:generic/example@1.0.0",
    visibility = ["//visibility:public"],
)

copyright_notice(
    name = "copyright_notice",
    text = [
        "Copyright © 2026 The authors of this stuff.",
        "Copyright © 1999 A. Cöntributor",
    ],
)
```

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.copyright_notice`, and its `attributes` file holds
the `text` entries as a JSON array of strings, encoded as UTF-8 and written
without insignificant whitespace:

```json
[
    "Copyright © 2026 The authors of this stuff.",
    "Copyright © 1999 A. Cöntributor"
]
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="copyright_notice-name"></a>name |  A unique name for this target.   |  none |
| <a id="copyright_notice-text"></a>text |  Required. One or more copyright notices. UTF-8 encoding.   |  none |
| <a id="copyright_notice-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="copyright_notice-visibility"></a>visibility |  The visibility of this target.   |  `None` |


