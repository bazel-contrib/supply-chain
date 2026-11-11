<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of `@package_metadata//licenses`.

<a id="LicenseExceptionInfo"></a>

## LicenseExceptionInfo

<pre>
load("@package_metadata//licenses:defs.bzl", "LicenseExceptionInfo")

LicenseExceptionInfo(<a href="#LicenseExceptionInfo-identifier">identifier</a>, <a href="#LicenseExceptionInfo-name">name</a>, *, <a href="#LicenseExceptionInfo-deprecation">deprecation</a>)
</pre>

Provides information to identify a license exception.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="LicenseExceptionInfo-identifier"></a>identifier | A [string](https://bazel.build/rules/lib/core/string) uniquely identifying the license exception (e.g., `Classpath-exception-2.0`). | none |
| <a id="LicenseExceptionInfo-name"></a>name | A [string](https://bazel.build/rules/lib/core/string) containing the (human readable) name of the license exception. | none |
| <a id="LicenseExceptionInfo-deprecation"></a>deprecation | An optional [string](https://bazel.build/rules/lib/core/string) explaining why the license exception is deprecated. | `None` |


<a id="LicenseKindInfo"></a>

## LicenseKindInfo

<pre>
load("@package_metadata//licenses:defs.bzl", "LicenseKindInfo")

LicenseKindInfo(<a href="#LicenseKindInfo-identifier">identifier</a>, <a href="#LicenseKindInfo-name">name</a>, *, <a href="#LicenseKindInfo-deprecation">deprecation</a>)
</pre>

Provides information to identify a license.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="LicenseKindInfo-identifier"></a>identifier | A [string](https://bazel.build/rules/lib/core/string) uniquely identifying the license (e.g., `Apache-2.0`, `EUPL-1.1`).<br><br>This is typically the [SPDX identifier](https://spdx.org/licenses/) of the license, but may also be a non-standard value (e.g., in case of a commercial license). | none |
| <a id="LicenseKindInfo-name"></a>name | A [string](https://bazel.build/rules/lib/core/string) containing the (human readable) name of the license (e.g., `Apache License 2.0`, `European Union Public License 1.1`) | none |
| <a id="LicenseKindInfo-deprecation"></a>deprecation | An optional [string](https://bazel.build/rules/lib/core/string) explaining why the license is deprecated (e.g., SPDX `GPL-2.0` in favor of `GPL-2.0-only`). | `None` |


<a id="license"></a>

## license

<pre>
load("@package_metadata//licenses:defs.bzl", "license")

license(*, <a href="#license-name">name</a>, <a href="#license-kind">kind</a>, <a href="#license-text">text</a>, <a href="#license-tags">tags</a>, <a href="#license-visibility">visibility</a>)
</pre>

Rule for declaring the license of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

Usage:

```starlark
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":license",
    ],
    purl = purl.builder().type("generic").name("example").version("1.0.0").build(),
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
    "expression": {
        "kind": {
            "identifier": "GPL-3.0-or-later",
            "name": "GNU General Public License v3.0 or later"
        }
    },
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
| `expression` | The root node of the license expression tree. Currently always a leaf: an object with a `kind`. Future versions add operator nodes (`AND`, `OR`, `WITH`) to this tree. |
| `expression.kind.identifier` | The `identifier` of the `kind` of this license. |
| `expression.kind.name` | The `full_name` of the `kind` of this license. |
| `expression.kind.custom` | `true` if the identifier is not on the SPDX license list. Absent otherwise. |
| `kind` | Deprecated, a copy of `expression.kind` for readers that do not know `expression` yet. |
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


<a id="license_exception"></a>

## license_exception

<pre>
load("@package_metadata//licenses:defs.bzl", "license_exception")

license_exception(*, <a href="#license_exception-name">name</a>, <a href="#license_exception-identifier">identifier</a>, <a href="#license_exception-full_name">full_name</a>, <a href="#license_exception-deprecation">deprecation</a>, <a href="#license_exception-tags">tags</a>, <a href="#license_exception-visibility">visibility</a>)
</pre>

Rule for declaring `LicenseExceptionInfo`.

A `license_exception` identifies an SPDX license exception, which is only
valid with a license in a `WITH` expression and is never a standalone
license.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="license_exception-name"></a>name |  A unique name for this target.   |  none |
| <a id="license_exception-identifier"></a>identifier |  Required. The unique identifier of the exception (e.g., `Classpath-exception-2.0`).   |  none |
| <a id="license_exception-full_name"></a>full_name |  Required. The (human readable) name of the exception.   |  none |
| <a id="license_exception-deprecation"></a>deprecation |  An optional message explaining why this target is deprecated.   |  `None` |
| <a id="license_exception-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="license_exception-visibility"></a>visibility |  The visibility of this target.   |  `None` |


<a id="license_kind"></a>

## license_kind

<pre>
load("@package_metadata//licenses:defs.bzl", "license_kind")

license_kind(*, <a href="#license_kind-name">name</a>, <a href="#license_kind-identifier">identifier</a>, <a href="#license_kind-full_name">full_name</a>, <a href="#license_kind-deprecation">deprecation</a>, <a href="#license_kind-tags">tags</a>, <a href="#license_kind-visibility">visibility</a>)
</pre>

Rule for declaring `LicenseKindInfo`.

A `license_kind` identifies a license; a `license` target refers to it to
declare that a package is available under that license.

Targets for all [SPDX licenses](https://spdx.org/licenses/) are predeclared
in `@package_metadata//licenses/spdx`, so declaring a `license_kind`
yourself is only necessary for licenses that are not in the SPDX list (e.g.,
a commercial license).

[SPDX license exceptions](https://spdx.org/licenses/exceptions-index.html)
are declared with the separate `license_exception` rule.

Usage:

```starlark
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")

license_kind(
    name = "acme_commercial",
    full_name = "ACME Commercial License 2.0",
    identifier = "LicenseRef-acme-commercial-2.0",
    visibility = ["//visibility:public"],
)
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="license_kind-name"></a>name |  A unique name for this target.   |  none |
| <a id="license_kind-identifier"></a>identifier |  Required. The unique identifier of the license (e.g., `Apache-2.0`, `EUPL-1.1`).<br><br>This is typically the [SPDX identifier](https://spdx.org/licenses/) of the license, but may also be a custom value (e.g., in case of a commercial license). It is not validated.<br><br>This identifies a single license; it is not an SPDX license expression (e.g., `MIT OR Apache-2.0`, or `GPL-2.0-only WITH Classpath-exception-2.0`).   |  none |
| <a id="license_kind-full_name"></a>full_name |  Required. The (human readable) name of the license (e.g., `Apache License 2.0`, `European Union Public License 1.1`).   |  none |
| <a id="license_kind-deprecation"></a>deprecation |  An optional message explaining why this target is deprecated.   |  `None` |
| <a id="license_kind-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="license_kind-visibility"></a>visibility |  The visibility of this target.   |  `None` |


