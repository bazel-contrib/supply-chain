<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `LicenseExceptionInfo`.

<a id="LicenseExceptionInfo"></a>

## LicenseExceptionInfo

<pre>
load("@package_metadata//licenses/providers:license_exception_info.bzl", "LicenseExceptionInfo")

LicenseExceptionInfo(<a href="#LicenseExceptionInfo-identifier">identifier</a>, <a href="#LicenseExceptionInfo-name">name</a>, *, <a href="#LicenseExceptionInfo-deprecation">deprecation</a>)
</pre>

Provides information to identify a license exception.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="LicenseExceptionInfo-identifier"></a>identifier | A [string](https://bazel.build/rules/lib/core/string) uniquely identifying the license exception (e.g., `Classpath-exception-2.0`). | none |
| <a id="LicenseExceptionInfo-name"></a>name | A [string](https://bazel.build/rules/lib/core/string) containing the (human readable) name of the license exception. | none |
| <a id="LicenseExceptionInfo-deprecation"></a>deprecation | An optional [string](https://bazel.build/rules/lib/core/string) explaining why the license exception is deprecated. | `None` |


