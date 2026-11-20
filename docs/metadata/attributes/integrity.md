<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `integrity`.

<a id="integrity"></a>

## integrity

<pre>
load("@package_metadata//attributes:integrity.bzl", "integrity")

integrity(*, <a href="#integrity-name">name</a>, <a href="#integrity-values">values</a>, <a href="#integrity-tags">tags</a>, <a href="#integrity-visibility">visibility</a>)
</pre>

Rule for declaring the integrity of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

Each integrity is a Subresource Integrity (SRI) string: the hash algorithm
name, a dash, then the base64-encoded digest, e.g. `sha256-<base64 digest>`.
The form and base64 encoding of each value are validated, and so is the size
of the digest of `sha256`, `sha384` and `sha512`. Values of other
algorithms are passed through as-is: tools consuming the attribute map them
to the format they produce, and may drop algorithms they do not support.

Entries are assumed to be UTF-8 encoded. Tools for building SBOMs presume
that.

Usage:

```starlark
load("@package_metadata//attributes:integrity.bzl", "integrity")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":integrity",
    ],
    purl = "pkg:maven/org.apache.commons/commons-lang3@3.12.0",
    visibility = ["//visibility:public"],
)

integrity(
    name = "integrity",
    values = [
        "sha384-/u6/iE9tq+bsqNaONz1r5IjNql63ZOiVKQM2/+n/lpaG8qnTYumou93257LhRV8t",
        "sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q=",
    ],
)
```

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.integrity`, and its `attributes` file holds the
`values` entries as a JSON array of strings, encoded as UTF-8 and
written without insignificant whitespace:

```json
[
    "sha384-/u6/iE9tq+bsqNaONz1r5IjNql63ZOiVKQM2/+n/lpaG8qnTYumou93257LhRV8t",
    "sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="
]
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="integrity-name"></a>name |  A unique name for this target.   |  none |
| <a id="integrity-values"></a>values |  Required. A list of SRI strings (`<algorithm>-<base64 digest>`) for the package, e.g. `["sha256-<base64 digest>"]`. Must contain at least one entry. UTF-8 encoding.   |  none |
| <a id="integrity-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="integrity-visibility"></a>visibility |  The visibility of this target.   |  `None` |


