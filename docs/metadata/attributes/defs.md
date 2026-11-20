<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Public API of `@package_metadata//attributes`.

<a id="copyright_notice"></a>

## copyright_notice

<pre>
load("@package_metadata//attributes:defs.bzl", "copyright_notice")

copyright_notice(*, <a href="#copyright_notice-name">name</a>, <a href="#copyright_notice-text">text</a>, <a href="#copyright_notice-tags">tags</a>, <a href="#copyright_notice-visibility">visibility</a>)
</pre>

Rule for declaring the copyright notice of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

Copyright `text` entries are assumed to be UTF-8 encoded. Tools for building
SBOMs presume that.

Usage:

```starlark
load("@package_metadata//attributes:copyright_notice.bzl", "copyright_notice")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":copyright_notice",
    ],
    purl = purl.builder().type("generic").name("example").version("1.0.0").build(),
    visibility = ["//visibility:public"],
)

copyright_notice(
    name = "copyright_notice",
    text = [
        "Copyright © 2026 The authors of this stuff.",
        "Copyright © 1999 A. Cöntributor",
    ],
)
```

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.copyright_notice`, and its `attributes` file holds
the `text` entries as a JSON array of strings, encoded as UTF-8 and written
without insignificant whitespace:

```json
[
    "Copyright © 2026 The authors of this stuff.",
    "Copyright © 1999 A. Cöntributor"
]
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="copyright_notice-name"></a>name |  A unique name for this target.   |  none |
| <a id="copyright_notice-text"></a>text |  Required. One or more copyright notices. UTF-8 encoding.   |  none |
| <a id="copyright_notice-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="copyright_notice-visibility"></a>visibility |  The visibility of this target.   |  `None` |


<a id="cpe"></a>

## cpe

<pre>
load("@package_metadata//attributes:defs.bzl", "cpe")

cpe(*, <a href="#cpe-name">name</a>, <a href="#cpe-identifiers">identifiers</a>, <a href="#cpe-tags">tags</a>, <a href="#cpe-visibility">visibility</a>)
</pre>

Rule for declaring the CPEs of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

`identifiers` entries are assumed to be UTF-8 encoded. Tools for building SBOMs
presume that.

Usage:

```starlark
load("@package_metadata//attributes:cpe.bzl", "cpe")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":cpe",
    ],
    purl = purl.builder().type("generic").name("ntp").version("4.2.8").build(),
    visibility = ["//visibility:public"],
)

cpe(
    name = "cpe",
    identifiers = [
        "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*",
    ],
)
```

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.cpe`, and its `attributes` file holds the `identifiers`
entries as a JSON array of strings, encoded as UTF-8 and written without
insignificant whitespace:

```json
[
    "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*"
]
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="cpe-name"></a>name |  A unique name for this target.   |  none |
| <a id="cpe-identifiers"></a>identifiers |  Required. One or more [CPE](https://en.wikipedia.org/wiki/Common_Platform_Enumeration) identifiers of the package. UTF-8 encoding.   |  none |
| <a id="cpe-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="cpe-visibility"></a>visibility |  The visibility of this target.   |  `None` |


<a id="integrity"></a>

## integrity

<pre>
load("@package_metadata//attributes:defs.bzl", "integrity")

integrity(*, <a href="#integrity-name">name</a>, <a href="#integrity-values">values</a>, <a href="#integrity-tags">tags</a>, <a href="#integrity-visibility">visibility</a>)
</pre>

Rule for declaring the integrity of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

Each integrity is a Subresource Integrity (SRI) string: the hash algorithm
name, a dash, then the base64-encoded digest, e.g. `sha256-<base64 digest>`.
Values are passed through as-is: this rule does not validate them. Tools
consuming the attribute are responsible for validating and mapping them to
the format they produce, and may drop algorithms they do not support.

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
        "sha1-he0IF6+DokrY2mjCtQlN5pgzmDw=",
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
    "sha1-he0IF6+DokrY2mjCtQlN5pgzmDw=",
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


<a id="originator"></a>

## originator

<pre>
load("@package_metadata//attributes:defs.bzl", "originator")

originator(*, <a href="#originator-name">name</a>, <a href="#originator-entity_name">entity_name</a>, <a href="#originator-type">type</a>, <a href="#originator-url">url</a>, <a href="#originator-email">email</a>, <a href="#originator-tags">tags</a>, <a href="#originator-visibility">visibility</a>)
</pre>

Declares who originally created a package.

The originator is the entity that originally created the package; it may
differ from the supplier that distributes it. This is typically passed to
the `attributes` of a `package_metadata` target.

Usage:

```starlark
load("@package_metadata//attributes:originator.bzl", "originator")

originator(
    name = "originator",
    entity_name = "A. Creator",
    type = "person",
    email = "creator@example",
)
```

Format: the `PackageAttributeInfo` kind is
`build.bazel.attribute.originator`; its JSON object has `name` and `type`,
and includes `url` and `email` only when set. CycloneDX maps this to the
component publisher string, so type, URL, and email are not represented.
SPDX 2.3 maps it to `PackageOriginator`, preserving type and email (in
parentheses) but not URL. SPDX 3.0.1 maps it to a `Person` or
`Organization` referenced by the package's `originatedBy`, preserving
email as an email external identifier and URL as an `altWebPage` external
reference.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="originator-name"></a>name |  A unique name for this target.   |  none |
| <a id="originator-entity_name"></a>entity_name |  Required. UTF-8 name of the original creator.   |  none |
| <a id="originator-type"></a>type |  `organization` or `person`; defaults to `organization`.   |  `"organization"` |
| <a id="originator-url"></a>url |  Optional URL for the originator.   |  `None` |
| <a id="originator-email"></a>email |  Optional email address for the originator.   |  `None` |
| <a id="originator-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="originator-visibility"></a>visibility |  The visibility of this target.   |  `None` |


<a id="patch"></a>

## patch

<pre>
load("@package_metadata//attributes:defs.bzl", "patch")

patch(*, <a href="#patch-name">name</a>, <a href="#patch-description">description</a>, <a href="#patch-integrity">integrity</a>, <a href="#patch-type">type</a>, <a href="#patch-url">url</a>, <a href="#patch-file">file</a>, <a href="#patch-resolves">resolves</a>, <a href="#patch-tags">tags</a>, <a href="#patch-visibility">visibility</a>)
</pre>

Rule for declaring a single patch applied to a package.

A patch is declared as its own target. Several of them can be passed to the
`attributes` of the same `package_metadata` target, since this attribute
sets `allow_multiple`; they are recorded in the order they are declared.

Values are passed through as-is: tools consuming the
attribute are responsible for mapping them to the format they produce.
Entries are assumed to be UTF-8 encoded.

Usage:

```starlark
load("@package_metadata//attributes:patch.bzl", "patch")

patch(
    name = "fix_overflow",
    description = "Fixes a buffer overflow in the parser.",
    integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
    resolves = ["CVE-2024-12345"],
    type = "backport",
    url = "https://example.com/patches/fix_overflow.patch",
)

patch(
    name = "local_tweak",
    description = "Local tweak.",
    file = "patches/local_tweak.patch",
    integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
    type = "unofficial",
)
```

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.patch`, and its `attributes` file holds a JSON object
written without insignificant whitespace, with keys sorted. `url`, `file`,
`file_exec_path` and `resolves` are omitted when unset. `file_exec_path` is
the path of `file` for tools that read its content:

```json
{
    "description": "Fixes a buffer overflow in the parser.",
    "integrity": ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
    "resolves": ["CVE-2024-12345"],
    "type": "backport",
    "url": "https://example.com/patches/fix_overflow.patch"
}
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="patch-name"></a>name |  A unique name for this target.   |  none |
| <a id="patch-description"></a>description |  Required. Human-readable description of what the patch does.   |  none |
| <a id="patch-integrity"></a>integrity |  Required. Checksums of the patch file as SRI strings (`<algorithm>-<base64 digest>`), e.g. `["sha256-<base64 digest>"]`. Must contain at least one entry.   |  none |
| <a id="patch-type"></a>type |  Required. The nature of the patch, one of `unofficial`, `monkey`, `backport` or `cherry-pick` (CycloneDX `pedigree.patches[].type`).   |  none |
| <a id="patch-url"></a>url |  Where the patch can be retrieved from. Exactly one of `url` or `file` must be set.   |  `None` |
| <a id="patch-file"></a>file |  A patch file in the workspace. Exactly one of `url` or `file` must be set. The JSON holds the path of the file relative to the root of the repository it belongs to, and the file is added to the `files` of the `PackageAttributeInfo`. `file_exec_path` holds its execution path.   |  `None` |
| <a id="patch-resolves"></a>resolves |  Optional. Identifiers (e.g., CVE ids, issue URLs) of the issues the patch resolves.   |  `None` |
| <a id="patch-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="patch-visibility"></a>visibility |  The visibility of this target.   |  `None` |


<a id="supplier"></a>

## supplier

<pre>
load("@package_metadata//attributes:defs.bzl", "supplier")

supplier(*, <a href="#supplier-name">name</a>, <a href="#supplier-entity_name">entity_name</a>, <a href="#supplier-type">type</a>, <a href="#supplier-url">url</a>, <a href="#supplier-email">email</a>, <a href="#supplier-tags">tags</a>, <a href="#supplier-visibility">visibility</a>)
</pre>

Declares who distributes a package.

The supplier is the entity that distributes the package, which can differ
from the entity that originally created it. This is typically passed to
the `attributes` of a `package_metadata` target.

Usage:

```starlark
load("@package_metadata//attributes:supplier.bzl", "supplier")

supplier(
    name = "supplier",
    entity_name = "ACME Corp",
    url = "https://acme.example",
    email = "oss@acme.example",
)
```

Format: the `PackageAttributeInfo` kind is
`build.bazel.attribute.supplier`; its JSON object has `name` and `type`,
and includes `url` and `email` only when set. CycloneDX maps this to the
component supplier entity, always as an organization even when `type` is
`person`; URL and email are preserved. SPDX 2.3 maps it to
`PackageSupplier`, preserving type and email (in parentheses) but not URL.
SPDX 3.0.1 maps it to a `Person` or `Organization` referenced by the
package's `suppliedBy`, preserving email as an email external identifier
and URL as an `altWebPage` external reference.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="supplier-name"></a>name |  A unique name for this target.   |  none |
| <a id="supplier-entity_name"></a>entity_name |  Required. UTF-8 name of the package distributor.   |  none |
| <a id="supplier-type"></a>type |  `organization` or `person`; defaults to `organization`.   |  `"organization"` |
| <a id="supplier-url"></a>url |  Optional URL for the supplier.   |  `None` |
| <a id="supplier-email"></a>email |  Optional email address for the supplier.   |  `None` |
| <a id="supplier-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="supplier-visibility"></a>visibility |  The visibility of this target.   |  `None` |


