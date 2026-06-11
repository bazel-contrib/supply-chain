<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `LicenseKindInfo`.

<a id="LicenseKindInfo"></a>

## LicenseKindInfo

<pre>
load("@package_metadata//licenses/providers:license_kind_info.bzl", "LicenseKindInfo")

LicenseKindInfo(<a href="#LicenseKindInfo-_init-identifier">identifier</a>, <a href="#LicenseKindInfo-_init-name">name</a>)
</pre>

Provides information to identify a license.

**CONSTRUCTOR PARAMETERS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="LicenseKindInfo-_init-identifier"></a>identifier | A [string](https://bazel.build/rules/lib/core/string) uniquely identifying the license (e.g., `Apache-2.0`, `EUPL-1.1`).<br><br>This is typically the [SPDX identifier](https://spdx.org/licenses/) of the license, but may also be a non-standard value (e.g., in case of a commercial license). | none |
| <a id="LicenseKindInfo-_init-name"></a>name | A [string](https://bazel.build/rules/lib/core/string) containing the (human readable) name of the license (e.g., `Apache License 2.0`, `European Union Public License 1.1`) | none |

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="LicenseKindInfo-kind"></a>kind |  A kind from the namespace of PackageAttributeInfo.    |
| <a id="LicenseKindInfo-identifier"></a>identifier |  A [string](https://bazel.build/rules/lib/core/string) uniquely identifying the license (e.g., `Apache-2.0`, `EUPL-1.1`).<br><br>This is typically the [SPDX identifier](https://spdx.org/licenses/) of the license, but may also be a non-standard value (e.g., in case of a commercial license).    |
| <a id="LicenseKindInfo-name"></a>name |  A [string](https://bazel.build/rules/lib/core/string) containing the (human readable) name of the license (e.g., `Apache License 2.0`, `European Union Public License 1.1`)    |


