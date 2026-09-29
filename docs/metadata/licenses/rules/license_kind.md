<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `license_kind`.

<a id="license_kind"></a>

## license_kind

<pre>
load("@package_metadata//licenses/rules:license_kind.bzl", "license_kind")

license_kind(*, <a href="#license_kind-name">name</a>, <a href="#license_kind-identifier">identifier</a>, <a href="#license_kind-full_name">full_name</a>, <a href="#license_kind-tags">tags</a>, <a href="#license_kind-visibility">visibility</a>)
</pre>

Rule for declaring `LicenseKindInfo`.

A `license_kind` identifies a license; a `license` target refers to it to
declare that a package is available under that license.

Targets for all [SPDX licenses](https://spdx.org/licenses/) are predeclared
in `@package_metadata//licenses/spdx`, so declaring a `license_kind`
yourself is only necessary for licenses that are not in the SPDX list (e.g.,
a commercial license).

[SPDX license exceptions](https://spdx.org/licenses/exceptions-index.html)
(e.g., `Classpath-exception-2.0`) are a separate list and are **not**
predeclared.

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
| <a id="license_kind-identifier"></a>identifier |  Required. The unique identifier of the license (e.g., `Apache-2.0`, `EUPL-1.1`).<br><br>This is typically the [SPDX identifier](https://spdx.org/licenses/) of the license, but may also be a non-standard value (e.g., in case of a commercial license). It is not validated.<br><br>This identifies a single license; it is not an SPDX license expression (e.g., `MIT OR Apache-2.0`, or `GPL-2.0-only WITH Classpath-exception-2.0`).   |  none |
| <a id="license_kind-full_name"></a>full_name |  Required. The (human readable) name of the license (e.g., `Apache License 2.0`, `European Union Public License 1.1`).   |  none |
| <a id="license_kind-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="license_kind-visibility"></a>visibility |  The visibility of this target.   |  `None` |


