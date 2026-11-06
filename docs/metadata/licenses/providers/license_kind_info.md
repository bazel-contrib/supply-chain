<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `LicenseKindInfo`.

<a id="LicenseKindInfo"></a>

## LicenseKindInfo

<pre>
load("@package_metadata//licenses/providers:license_kind_info.bzl", "LicenseKindInfo")

LicenseKindInfo(<a href="#LicenseKindInfo-identifier">identifier</a>, <a href="#LicenseKindInfo-name">name</a>, *, <a href="#LicenseKindInfo-deprecated">deprecated</a>, <a href="#LicenseKindInfo-exception">exception</a>)
</pre>

Provides information to identify a license.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="LicenseKindInfo-identifier"></a>identifier | A [string](https://bazel.build/rules/lib/core/string) uniquely identifying the license (e.g., `Apache-2.0`, `EUPL-1.1`).<br><br>This is typically the [SPDX identifier](https://spdx.org/licenses/) of the license, but may also be a non-standard value (e.g., in case of a commercial license). | none |
| <a id="LicenseKindInfo-name"></a>name | A [string](https://bazel.build/rules/lib/core/string) containing the (human readable) name of the license (e.g., `Apache License 2.0`, `European Union Public License 1.1`) | none |
| <a id="LicenseKindInfo-deprecated"></a>deprecated | A [bool](https://bazel.build/rules/lib/core/bool) indicating that the license identifier is deprecated (e.g., SPDX `GPL-2.0` in favor of `GPL-2.0-only`). | `False` |
| <a id="LicenseKindInfo-exception"></a>exception | A [bool](https://bazel.build/rules/lib/core/bool) indicating that this is a license exception (e.g., `Classpath-exception-2.0`), which is only valid on the right-hand side of `WITH`, rather than a license. | `False` |


