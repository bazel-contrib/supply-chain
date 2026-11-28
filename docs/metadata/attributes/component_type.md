<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `component_type`.

<a id="component_type"></a>

## component_type

<pre>
load("@package_metadata//attributes:component_type.bzl", "component_type")

component_type(*, <a href="#component_type-name">name</a>, <a href="#component_type-type">type</a>, <a href="#component_type-tags">tags</a>, <a href="#component_type-visibility">visibility</a>)
</pre>

Rule for declaring the component type of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

The type tells SBOM consumers what kind of thing the package is. Tools for
building SBOMs report it as the CycloneDX `component.type`, the SPDX 2.3
`primaryPackagePurpose` and the SPDX 3.0.1 `software_primaryPurpose`. Only the
types supported by all of these formats are accepted. When a package does not
declare one, CycloneDX reports `library` and the SPDX formats report no
purpose.

Usage:

```starlark
load("@package_metadata//attributes:component_type.bzl", "component_type")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":component_type",
    ],
    purl = "pkg:oci/my-image?repository_url=registry.example.com%2Fmy-image",
    visibility = ["//visibility:public"],
)

component_type(
    name = "component_type",
    type = "container",
)
```

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.component_type`, and its `attributes` file holds the
`type` as a JSON string, written without insignificant whitespace:

```json
"container"
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="component_type-name"></a>name |  A unique name for this target.   |  none |
| <a id="component_type-type"></a>type |  Required. The type of the component, one of `application`, `container`, `device`, `file`, `firmware`, `framework`, `library` or `operating-system`.   |  none |
| <a id="component_type-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="component_type-visibility"></a>visibility |  The visibility of this target.   |  `None` |


