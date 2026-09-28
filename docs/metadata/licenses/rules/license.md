<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `license`.

<a id="license"></a>

## license

<pre>
load("@package_metadata//licenses/rules:license.bzl", "license")

license(*, <a href="#license-name">name</a>, <a href="#license-kind">kind</a>, <a href="#license-text">text</a>, <a href="#license-tags">tags</a>, <a href="#license-visibility">visibility</a>)
</pre>

Rule for declaring the license of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

Usage:

```starlark
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":license",
    ],
    purl = "pkg:generic/example@1.0.0",
    visibility = ["//visibility:public"],
)

license(
    name = "license",
    kind = "@package_metadata//licenses/spdx:GPL-3.0-or-later",
    text = "COPYING",
)
```

A `license` target declares exactly one license kind. Since the
`attributes` of a `package_metadata` target are keyed by `kind`, at most one
`license` attribute per `package_metadata` target is meaningful; declare
separate `package_metadata` targets for parts of a package that are under a
different license. Packages available under a choice of licenses (e.g.,
`MIT OR Apache-2.0`) and SPDX license exceptions are not modelled yet.

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.license`, and its `attributes` file holds a JSON
object, encoded as UTF-8 and written without insignificant whitespace (the
example below is indented for readability):

```json
{
    "kind": {
        "identifier": "GPL-3.0-or-later",
        "name": "GNU General Public License v3.0 or later"
    },
    "label": "@@//:license",
    "text": "COPYING"
}
```

| Field | Description |
| :---- | :---------- |
| `kind.identifier` | The `identifier` of the `kind` of this license. |
| `kind.name` | The `full_name` of the `kind` of this license. |
| `label` | The label of this `license` target. |
| `text` | The path of the license `text` file, relative to the execution root. Absent if no `text` was declared. |


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="license-name"></a>name |  A unique name for this target.   |  none |
| <a id="license-kind"></a>kind |  Required. The kind of license this license is classified as.<br><br>This is typically a `license_kind` target. Targets for all SPDX licenses are predeclared in `@package_metadata//licenses/spdx`.   |  none |
| <a id="license-text"></a>text |  The [File](https://bazel.build/rules/lib/builtins/File) with the text of the license.<br><br>This is typically the `LICENSE` or `COPYING` file of the package. It is propagated to consumers so that it can be shipped alongside built artifacts.   |  `None` |
| <a id="license-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="license-visibility"></a>visibility |  The visibility of this target.   |  `None` |


