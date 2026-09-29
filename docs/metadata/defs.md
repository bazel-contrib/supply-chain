<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of `@package_metadata`.

<a id="PackageAttributeInfo"></a>

## PackageAttributeInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageAttributeInfo")

PackageAttributeInfo(<a href="#PackageAttributeInfo-kind">kind</a>, <a href="#PackageAttributeInfo-attributes">attributes</a>, <a href="#PackageAttributeInfo-files">files</a>)
</pre>

Provider for declaring a single attribute of a Bazel package (e.g., the license
it is available under or the copyright notice).

Attributes are attached to a package by passing targets providing this to the
`attributes` of a `package_metadata` target. Declaring one is the extension
point for organizations that need to inject metadata beyond the attributes
provided by `@package_metadata`.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageAttributeInfo-kind"></a>kind | The identifier of the attribute.<br><br>This should generally be in reverse DNS format (e.g., `com.example.foo`). | none |
| <a id="PackageAttributeInfo-attributes"></a>attributes | The [File](https://bazel.build/rules/lib/builtins/File) containing the attributes.<br><br>The format of this file depends on the `kind` of attribute. Please consult the documentation of the attribute (e.g., `license` documents the JSON object it writes). | none |
| <a id="PackageAttributeInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s containing information about this attribute. | `[]` |


<a id="PackageMetadataInfo"></a>

## PackageMetadataInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageMetadataInfo")

PackageMetadataInfo(<a href="#PackageMetadataInfo-metadata">metadata</a>, <a href="#PackageMetadataInfo-files">files</a>)
</pre>

Provider for declaring metadata about a Bazel package.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataInfo-metadata"></a>metadata | The [File](https://bazel.build/rules/lib/builtins/File) containing metadata about the package. | none |
| <a id="PackageMetadataInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the package, including transitive files from all attributes of the package. | `[]` |


<a id="PackageMetadataOverrideInfo"></a>

## PackageMetadataOverrideInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageMetadataOverrideInfo")

PackageMetadataOverrideInfo(*, <a href="#PackageMetadataOverrideInfo-packages">packages</a>, <a href="#PackageMetadataOverrideInfo-metadata">metadata</a>)
</pre>

Defines an override for `PackageMetadataInfo` for a set of packages.

This is typically used to attach metadata to a dependency that does not declare
any itself.

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="PackageMetadataOverrideInfo-packages"></a>packages | A [PackageSpecificationInfo](https://bazel.build/rules/lib/providers/PackageSpecificationInfo) provider declaring which packages the override applies to.<br><br>This is typically created by a [package_group](https://bazel.build/rules/lib/globals/build#package_group) target. |
| <a id="PackageMetadataOverrideInfo-metadata"></a>metadata | The `PackageMetadataInfo` provider to use instead of the provider declared by package itself. |


<a id="PackageMetadataToolchainInfo"></a>

## PackageMetadataToolchainInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageMetadataToolchainInfo")

PackageMetadataToolchainInfo(<a href="#PackageMetadataToolchainInfo-metadata_overrides">metadata_overrides</a>)
</pre>

Toolchain for `package_metadata`.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataToolchainInfo-metadata_overrides"></a>metadata_overrides | A sequence of `PackageMetadataOverrideInfo` providers. | `[]` |


<a id="TargetInfo"></a>

## TargetInfo

<pre>
load("@package_metadata//:defs.bzl", "TargetInfo")

TargetInfo(<a href="#TargetInfo-_init-metadata">metadata</a>, <a href="#TargetInfo-_init-files">files</a>)
</pre>

Provider for describing a single target.

This includes the `PackageMetadataInfo`s directly attached to the target as well
as `TargetInfo` from dependencies of the target.

`TargetInfo` provides information about a single node in the
(configured) target graph, including outgoing edges to its direct dependencies.

**CONSTRUCTOR PARAMETERS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="TargetInfo-_init-metadata"></a>metadata | <p align="center">-</p> | none |
| <a id="TargetInfo-_init-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the target, including transitive files from all dependencies. | `[]` |

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="TargetInfo-files"></a>files |  A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s with metadata about the target, including transitive files from all dependencies.    |
| <a id="TargetInfo-info"></a>info |  The [File](https://bazel.build/rules/lib/builtins/File) containing the information about the target.    |


<a id="package_metadata"></a>

## package_metadata

<pre>
load("@package_metadata//:defs.bzl", "package_metadata")

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


<a id="package_metadata_common.create_package_metadata"></a>

## package_metadata_common.create_package_metadata

<pre>
load("@package_metadata//:defs.bzl", "package_metadata_common")

package_metadata_common.create_package_metadata(*, <a href="#package_metadata_common.create_package_metadata-actions">actions</a>, <a href="#package_metadata_common.create_package_metadata-label">label</a>, <a href="#package_metadata_common.create_package_metadata-purl">purl</a>, <a href="#package_metadata_common.create_package_metadata-attributes">attributes</a>)
</pre>

Creates a PackageMetadataInfo provider with JSON metadata.

This function generates a JSON file containing metadata about a Bazel package,
including its PURL (Package URL), label, and attributes. The metadata is
structured for consumption by supply chain analysis tools.

**Example:**

```starlark
def _my_rule_impl(ctx):
    info = create_package_metadata(
        actions = ctx.actions,
        label = ctx.label,
        purl = "pkg:npm/my-package@1.0.0",
        attributes = [a[PackageAttributeInfo] for a in ctx.attr.attributes],
    )
    return [info]
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata_common.create_package_metadata-actions"></a>actions |  The [actions](https://bazel.build/rules/lib/builtins/actions) object from the rule context, used to declare and write files.   |  none |
| <a id="package_metadata_common.create_package_metadata-label"></a>label |  The [Label](https://bazel.build/rules/lib/builtins/Label) of the target being processed.   |  none |
| <a id="package_metadata_common.create_package_metadata-purl"></a>purl |  A string containing the [PURL](https://github.com/package-url/purl-spec) uniquely identifying this package (e.g., "pkg:npm/lodash@4.17.21").   |  none |
| <a id="package_metadata_common.create_package_metadata-attributes"></a>attributes |  A list of [PackageAttributeInfo](//providers:package_attribute_info.bzl) providers representing package attributes (e.g., source location, license). Defaults to an empty list.   |  `[]` |

**RETURNS**

A [PackageMetadataInfo](//providers:package_metadata_info.bzl) provider
  containing the generated metadata file and transitive files from all
  attributes.


<a id="package_metadata_common.create_target_info"></a>

## package_metadata_common.create_target_info

<pre>
load("@package_metadata//:defs.bzl", "package_metadata_common")

package_metadata_common.create_target_info(*, <a href="#package_metadata_common.create_target_info-actions">actions</a>, <a href="#package_metadata_common.create_target_info-label">label</a>, <a href="#package_metadata_common.create_target_info-package_metadata">package_metadata</a>)
</pre>

Creates a TargetInfo provider with JSON metadata.

This function generates a JSON file containing metadata about a Bazel target,
including its label and references to package metadata from its dependencies.
The metadata is structured for consumption by supply chain analysis tools.

**Example:**

```starlark
def _my_rule_impl(ctx):
    info = create_target_info(
        actions = ctx.actions,
        label = ctx.label,
        package_metadata = [
            dep[PackageMetadataInfo]
            for dep in ctx.attr.deps
            if PackageMetadataInfo in dep
        ],
    )
    return [info]
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata_common.create_target_info-actions"></a>actions |  The [actions](https://bazel.build/rules/lib/builtins/actions) object from the rule context, used to declare and write files.   |  none |
| <a id="package_metadata_common.create_target_info-label"></a>label |  The [Label](https://bazel.build/rules/lib/builtins/Label) of the target being processed.   |  none |
| <a id="package_metadata_common.create_target_info-package_metadata"></a>package_metadata |  A list of [PackageMetadataInfo](//providers:package_metadata_info.bzl) providers directly attached to the target being processed Defaults to an empty list.   |  `[]` |

**RETURNS**

A [TargetInfo](//providers:target_info.bzl) provider
  containing the generated metadata file and transitive files from all
  package metadata.


<a id="purl.bazel"></a>

## purl.bazel

<pre>
load("@package_metadata//:defs.bzl", "purl")

purl.bazel(<a href="#purl.bazel-name">name</a>, <a href="#purl.bazel-version">version</a>, <a href="#purl.bazel-registry">registry</a>)
</pre>

Defines a `purl` for a Bazel module.

This is typically used to construct `purl` for `package_metadata` targets in
Bazel modules.

This is **NOT** supported in `WORKSPACE` mode.

Example:

```starlark
load("@package_metadata//purl:purl.bzl", "purl")

package_metadata(
    name = "package_metadata",
    purl = purl.bazel(module_name(), module_version()),
    attributes = [
        # ...
    ],
    visibility = ["//visibility:public"],
)
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="purl.bazel-name"></a>name |  The name of the Bazel module. Typically [module_name()](https://bazel.build/rules/lib/globals/build#module_name).   |  none |
| <a id="purl.bazel-version"></a>version |  The version of the Bazel module. Typically [module_version()](https://bazel.build/rules/lib/globals/build#module_version). May be empty or `None`.   |  none |
| <a id="purl.bazel-registry"></a>registry |  The URL of the registry that hosts the Bazel module. Defaults to https://bcr.bazel.build.   |  `"https://bcr.bazel.build"` |

**RETURNS**

The `purl` for the Bazel module (e.g. `pkg:bazel/foo` or
  `pkg:bazel/bar@1.2.3`).


<a id="purl.builder"></a>

## purl.builder

<pre>
load("@package_metadata//:defs.bzl", "purl")

purl.builder()
</pre>

Creates a fluent builder for constructing Package URLs (PURLs).

The builder provides a chainable interface for constructing PURLs according to
the [Package URL specification](https://github.com/package-url/purl-spec).

The `type` and `name` fields are required. All components are validated and
normalized according to the PURL spec. Components are automatically percent-encoded
where necessary, and qualifiers are sorted lexicographically in the output.

For a list of supported PURL types and their specifications, see:
https://github.com/package-url/purl-spec/blob/main/purl-types-index.json

Example - Simple PURL:

```starlark
load("@package_metadata//purl:purl.bzl", "purl")

my_purl = (purl.builder()
    .type("npm")
    .name("foobar")
    .version("12.3.1")
    .build())
# Result: pkg:npm/foobar@12.3.1
```

Example - Maven with namespace and qualifiers:

```starlark
load("@package_metadata//purl:purl.bzl", "purl")

my_purl = (purl.builder()
    .type("maven")
    .namespace("org.apache.xmlgraphics")
    .name("batik-anim")
    .version("1.9.1")
    .add_qualifier("classifier", "sources")
    .add_qualifier("repository_url", "https://repo.spring.io/release")
    .build())
# Result: pkg:maven/org.apache.xmlgraphics/batik-anim@1.9.1?classifier=sources&repository_url=https%3A%2F%2Frepo.spring.io%2Frelease
```

Example - Golang with namespace and subpath:

```starlark
load("@package_metadata//purl:purl.bzl", "purl")

my_purl = (purl.builder()
    .type("golang")
    .namespace("google.golang.org")
    .name("genproto")
    .version("abcdedf")
    .subpath("googleapis/api/annotations")
    .build())
# Result: pkg:golang/google.golang.org/genproto@abcdedf#googleapis/api/annotations
```



**RETURNS**

A builder object with chainable methods:

  - `type(type_name)`: Sets the package type (required). Must be lowercase ASCII.
  - `namespace(namespace)`: Sets the namespace (optional). String with segments separated by '/'.
  - `name(name)`: Sets the package name (required).
  - `version(version)`: Sets the package version (optional).
  - `add_qualifier(name, value)`: Adds a qualifier (optional, repeatable).
    Key must start with ASCII letter and contain only lowercase letters,
    numbers, '.', '-', '_'.
  - `subpath(subpath)`: Sets the subpath (optional). String with segments separated by '/'.
  - `disable_checks()`: Disables validation and normalization of the PURL.
  - `build()`: Validates, normalizes, and constructs the final PURL string.
    Performs both general and type-specific validation and normalization.
    Fails if validation errors occur.


<a id="purl.parse"></a>

## purl.parse

<pre>
load("@package_metadata//:defs.bzl", "purl")

purl.parse(<a href="#purl.parse-value">value</a>)
</pre>

Parses a PURL string into normalized components.

The parsing flow implements ECMA-427 1st edition, December 2025,
§5.6 "Rules for each PURL component".

See https://ecma-international.org/wp-content/uploads/ECMA-427_1st_edition_december_2025.pdf

It parses the components in reverse order of their appearance in the PURL
string: parsing right to left avoids ambiguity between the separators of the
components.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="purl.parse-value"></a>value |  The PURL string to parse.   |  none |

**RETURNS**

A tuple of (purl_components, error). On success, error is `None` and
  `purl_components` is a [dict](https://bazel.build/rules/lib/core/dict)
  with the normalized, percent-decoded components of the PURL:

  - `type`: The package type (e.g., `npm`). Always present.
  - `namespace`: The namespace, with segments joined by '/' (e.g.,
    `org.apache.xmlgraphics`), or `None`.
  - `name`: The package name. Always present.
  - `version`: The version (e.g., `1.9.1`), or `None`.
  - `qualifiers`: A dict of qualifier key-value pairs, or `None`.
  - `subpath`: The subpath, with segments joined by '/' (e.g.,
    `googleapis/api/annotations`), or `None`.

  On failure, `purl_components` is `None` and `error` is a message
  describing why `value` is not a valid PURL.


