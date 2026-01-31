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





