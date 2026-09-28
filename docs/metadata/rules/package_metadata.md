<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `package_metadata`.

<a id="package_metadata"></a>

## package_metadata

<pre>
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(*, <a href="#package_metadata-name">name</a>, <a href="#package_metadata-purl">purl</a>, <a href="#package_metadata-attributes">attributes</a>, <a href="#package_metadata-visibility">visibility</a>, <a href="#package_metadata-tags">tags</a>)
</pre>

Rule for declaring `PackageMetadataInfo`, typically of a `bzlmod` module.

A `package_metadata` target identifies a package and where it was retrieved
from. Additional metadata (e.g., the license the package is under) is
attached by passing `attributes`.

Targets are associated with a `package_metadata` target module-wide via
`repo(default_package_metadata = ...)` in `REPO.bazel`, package-wide via
`package(default_package_metadata = ...)`, or individually via the
`package_metadata` attribute of the target.

Usage:

```starlark
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":license",
    ],
    purl = purl.bazel(module_name(), module_version()),
    visibility = ["//visibility:public"],
)
```

Format: the `metadata` file of the `PackageMetadataInfo` of this rule holds
a JSON object, encoded as UTF-8 and written without insignificant whitespace
(the example below is indented for readability):

```json
{
    "attributes": {
        "build.bazel.attribute.license": "bazel-out/k8-fastbuild/bin/license.package-attribute.json"
    },
    "label": "@@//:package_metadata",
    "purl": "pkg:bazel/package_metadata"
}
```

| Field | Description |
| :---- | :---------- |
| `attributes` | The declared `attributes`, keyed by the `kind` of their `PackageAttributeInfo`. Values are the paths of the `attributes` files, relative to the execution root. |
| `label` | The label of this `package_metadata` target. |
| `purl` | The `purl` of this target, verbatim. |

Since `attributes` is keyed by `kind`, declaring two attributes of the same
`kind` on one target is not meaningful — only one of them is recorded.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata-name"></a>name |  A unique name for this target.   |  none |
| <a id="package_metadata-purl"></a>purl |  Required. The [PURL](https://github.com/package-url/purl-spec) uniquely identifying this package.<br><br>For Bazel modules, this is typically constructed with `purl.bazel`.   |  none |
| <a id="package_metadata-attributes"></a>attributes |  A list of `attributes` of the package (e.g., source location, license, ...).<br><br>Each element must be a target providing `PackageAttributeInfo` (e.g., a `license` target).   |  `[]` |
| <a id="package_metadata-visibility"></a>visibility |  The visibility of this target.<br><br>`package_metadata` targets are typically `//visibility:public` so that consumers of the module can read the metadata.   |  `None` |
| <a id="package_metadata-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |


