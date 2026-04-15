<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of `@package_metadata`.

<a id="PackageAttributeInfo"></a>

## PackageAttributeInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageAttributeInfo")

PackageAttributeInfo(<a href="#PackageAttributeInfo-kind">kind</a>, <a href="#PackageAttributeInfo-attributes">attributes</a>, <a href="#PackageAttributeInfo-files">files</a>)
</pre>

Provider for declaring metadata about a Bazel package.

> **Fields in this provider are not covered by the stability guarantee.**

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageAttributeInfo-kind"></a>kind | The identifier of the attribute.<br><br>This should generally be in reverse DNS format (e.g., `com.example.foo`). | none |
| <a id="PackageAttributeInfo-attributes"></a>attributes | The [File](https://bazel.build/rules/lib/builtins/File) containing the attributes.<br><br>The format of this file depends on the `kind` of attribute. Please consult the documentation of the attribute. | none |
| <a id="PackageAttributeInfo-files"></a>files | A [depset](https://bazel.build/rules/lib/builtins/depset) of [File](https://bazel.build/rules/lib/builtins/File)s containing information about this attribute. | `[]` |


<a id="PackageMetadataInfo"></a>

## PackageMetadataInfo

<pre>
load("@package_metadata//:defs.bzl", "PackageMetadataInfo")

PackageMetadataInfo(<a href="#PackageMetadataInfo-metadata">metadata</a>, <a href="#PackageMetadataInfo-files">files</a>)
</pre>

Provider for declaring metadata about a Bazel package.

> **Fields in this provider are not covered by the stability guarantee.**

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

> **Fields in this provider are not covered by the stability guarantee.**

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

> **Fields in this provider are not covered by the stability guarantee.**

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="PackageMetadataToolchainInfo-metadata_overrides"></a>metadata_overrides | A sequence of `PackageMetadataOverrideInfo` providers. | `[]` |


<a id="package_metadata"></a>

## package_metadata

<pre>
load("@package_metadata//:defs.bzl", "package_metadata")

package_metadata(*, <a href="#package_metadata-name">name</a>, <a href="#package_metadata-purl">purl</a>, <a href="#package_metadata-attributes">attributes</a>, <a href="#package_metadata-visibility">visibility</a>, <a href="#package_metadata-tags">tags</a>)
</pre>



**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="package_metadata-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="package_metadata-purl"></a>purl |  <p align="center"> - </p>   |  none |
| <a id="package_metadata-attributes"></a>attributes |  <p align="center"> - </p>   |  `[]` |
| <a id="package_metadata-visibility"></a>visibility |  <p align="center"> - </p>   |  `None` |
| <a id="package_metadata-tags"></a>tags |  <p align="center"> - </p>   |  `None` |


<a id="purl.bazel"></a>

## purl.bazel

<pre>
load("@package_metadata//:defs.bzl", "purl")

purl.bazel(<a href="#purl.bazel-name">name</a>, <a href="#purl.bazel-version">version</a>)
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

    load("@package_metadata//purl:purl.bzl", "purl")

    my_purl = (purl.builder()
        .type("npm")
        .name("foobar")
        .version("12.3.1")
        .build())
    # Result: pkg:npm/foobar@12.3.1

Example - Maven with namespace and qualifiers:

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

Example - Golang with namespace and subpath:

    load("@package_metadata//purl:purl.bzl", "purl")

    my_purl = (purl.builder()
        .type("golang")
        .namespace("google.golang.org")
        .name("genproto")
        .version("abcdedf")
        .subpath("googleapis/api/annotations")
        .build())
    # Result: pkg:golang/google.golang.org/genproto@abcdedf#googleapis/api/annotations



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
  - `build()`: Validates, normalizes, and constructs the final PURL string.
    Performs both general and type-specific validation and normalization.
    Fails if validation errors occur.


