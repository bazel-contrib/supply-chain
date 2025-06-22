<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of `@package_metadata//licenses`.

<a id="LicenseKindInfo"></a>

## LicenseKindInfo

<pre>
load("@package_metadata//licenses:defs.bzl", "LicenseKindInfo")

LicenseKindInfo(<a href="#LicenseKindInfo-identifier">identifier</a>, <a href="#LicenseKindInfo-name">name</a>)
</pre>

Provides information to identify a license.

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="LicenseKindInfo-identifier"></a>identifier | A [string](https://bazel.build/rules/lib/core/string) uniquely identifying the license (e.g., `Apache-2.0`, `EUPL-1.1`).<br><br>This is typically the [SPDX identifier](https://spdx.org/licenses/) of the license, but may also be a non-standard value (e.g., in case of a commercial license). |
| <a id="LicenseKindInfo-name"></a>name | A [string](https://bazel.build/rules/lib/core/string) containing the (human readable) name of the license (e.g., `Apache License 2.0`, `European Union Public License 1.1`) |


<a id="license_kind"></a>

## license_kind

<pre>
load("@package_metadata//licenses:defs.bzl", "license_kind")

license_kind(*, <a href="#license_kind-name">name</a>, <a href="#license_kind-identifier">identifier</a>, <a href="#license_kind-full_name">full_name</a>, <a href="#license_kind-visibility">visibility</a>)
</pre>



**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="license_kind-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="license_kind-identifier"></a>identifier |  <p align="center"> - </p>   |  none |
| <a id="license_kind-full_name"></a>full_name |  <p align="center"> - </p>   |  none |
| <a id="license_kind-visibility"></a>visibility |  <p align="center"> - </p>   |  `None` |


