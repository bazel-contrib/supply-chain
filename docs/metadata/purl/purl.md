<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Module defining urils for [purl](https://github.com/package-url/purl-spec)s.

<a id="purl.bazel"></a>

## purl.bazel

<pre>
load("@package_metadata//purl:purl.bzl", "purl")

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
load("@package_metadata//purl:purl.bzl", "purl")

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


