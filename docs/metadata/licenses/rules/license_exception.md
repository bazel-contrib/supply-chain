<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `license_exception`.

<a id="license_exception"></a>

## license_exception

<pre>
load("@package_metadata//licenses/rules:license_exception.bzl", "license_exception")

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


