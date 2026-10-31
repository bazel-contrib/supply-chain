<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `license`.

<a id="license"></a>

## license

<pre>
load("@package_metadata//licenses/rules:license.bzl", "license")

license(*, <a href="#license-name">name</a>, <a href="#license-kind">kind</a>, <a href="#license-expression">expression</a>, <a href="#license-kinds">kinds</a>, <a href="#license-text">text</a>, <a href="#license-tags">tags</a>, <a href="#license-visibility">visibility</a>)
</pre>

Rule for declaring the license of a package or target.

This is typically passed to the `attributes` of a `package_metadata` target.

Usage:

```starlark
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    attributes = [
        ":license",
    ],
    purl = "pkg:generic/example@1.0.0",
    visibility = ["//visibility:public"],
)

license(
    name = "license",
    kind = "@package_metadata//licenses/spdx:GPL-3.0-or-later",
    text = "COPYING",
)
```

A `license` target declares exactly one license kind. Since the
`attributes` of a `package_metadata` target are keyed by `kind`, at most one
`license` attribute per `package_metadata` target is meaningful; declare
separate `package_metadata` targets for parts of a package that are under a
different license. Packages available under a choice of licenses (e.g.,
`MIT OR Apache-2.0`) and SPDX license exceptions can be declared with
`expression` instead of `kind`. The parsed expression is a graph of
`LicenseKindInfo`s, also exposed through the `LicenseExpressionInfo`
provider of the `license` target. For a single `kind`, that provider is a
leaf node holding the `LicenseKindInfo`.

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.license`, and its `attributes` file holds a JSON
object, encoded as UTF-8 and written without insignificant whitespace (the
example below is indented for readability):

A single license, e.g. `kind = "@package_metadata//licenses/spdx:GPL-3.0-or-later"`,
is a leaf node of `expression`. For backward compatibility with tools that
predate `expression`, the same kind is also written to the deprecated
top-level `kind`:

```json
{
    "expression": {
        "kind": {
            "identifier": "GPL-3.0-or-later",
            "name": "GNU General Public License v3.0 or later"
        }
    },
    "kind": {
        "identifier": "GPL-3.0-or-later",
        "name": "GNU General Public License v3.0 or later"
    },
    "label": "@@//:license",
    "text": "COPYING"
}
```

An `expression` such as `MIT OR (GPL-2.0+ AND Apache-2.0 WITH LLVM-exception)`
is written as a tree of operator nodes whose operands are nodes. A leaf can
carry `or_later` (from `+`), and a `WITH` node has a single leaf operand
and an `exception`. As older tools cannot represent this, the deprecated
top-level `kind` is `null` (it is also `null` for a single license followed
by `+`):

```json
{
    "expression": {
        "operator": "OR",
        "operands": [
            {
                "kind": {"identifier": "MIT", "name": "MIT License"}
            },
            {
                "operator": "AND",
                "operands": [
                    {
                        "kind": {"identifier": "GPL-2.0-only", "name": "GNU General Public License v2.0 only"},
                        "or_later": true
                    },
                    {
                        "operator": "WITH",
                        "operands": [
                            {"kind": {"identifier": "Apache-2.0", "name": "Apache License 2.0"}}
                        ],
                        "exception": {"identifier": "LLVM-exception", "name": "LLVM Exception"}
                    }
                ]
            }
        ]
    },
    "kind": null,
    "label": "@@//:license"
}
```

Grammar of the `expression` string, from the [SPDX 3.0.1 license expressions specification](https://spdx.github.io/spdx-spec/v3.0.1/annexes/spdx-license-expressions/) (ABNF, RFC 5234 and RFC 7405):

```
idstring = 1*(ALPHA / DIGIT / "-" / "." )
license-id = <short form license identifier from SPDX License List>
license-exception-id = <short form license exception identifier from SPDX License List>
license-ref = [%s"DocumentRef-"(idstring)":"]%s"LicenseRef-"(idstring)
addition-ref = [%s"DocumentRef-"(idstring)":"]%s"AdditionRef-"(idstring)
simple-expression = license-id / license-id"+" / license-ref
addition-expression = license-exception-id / addition-ref
compound-expression = (simple-expression /
  simple-expression ( %s"WITH" / %s"with" ) addition-expression /
  compound-expression ( %s"AND" / %s"and" ) compound-expression /
  compound-expression ( %s"OR" / %s"or" ) compound-expression /
  "(" compound-expression ")" )
license-expression = (simple-expression / compound-expression)
```

As in SPDX, `+` binds tighter than `WITH`, which binds tighter than `AND`,
which binds tighter than `OR`. There must be no whitespace between a
license id and `+`, and whitespace around `WITH`. `AND` and `OR` need
whitespace and/or parentheses around them. In addition, this parser:

* matches `license-id` and `license-exception-id` case-insensitively, as
  SPDX recommends, and the part of a reference after `LicenseRef-` or
  `AdditionRef-` as well;
* does not yet accept the lowercase operators `and`, `or` and `with`, nor a
  `DocumentRef-` prefix on an `addition-ref`;
* requires every `license-ref` and `addition-ref` to be declared as a
  `license_kind` and passed in `kinds`.

The shape of `expression`, in ABNF-like notation (the ids are SPDX
license ids, see the [SPDX specification](https://spdx.github.io/spdx-spec/v3.0.1/annexes/spdx-license-expressions/)):

```
node     = leaf / with / compound
leaf     = { "kind": kind [, "or_later": true] }
with     = { "operator": "WITH", "operands": [ leaf ], "exception": kind }
compound = { "operator": "AND" / "OR", "operands": 2*node }
kind     = { "identifier": string, "name": string [, "custom": true] }
```

`WITH` binds tighter than `AND`, which binds tighter than `OR`. Nested
nodes with the same operator are flattened into one node with more
operands, and operands are not de-duplicated.

| Field | Description |
| :---- | :---------- |
| `expression` | The root node of the AST of the license. Always present: a single `kind` is a leaf node, and an SPDX license `expression` is parsed into a tree. A node is either a leaf `{"kind": {"identifier", "name", "custom"}, "or_later"}` or `{"operator": "AND" or "OR" or "WITH", "operands": [...], "exception"}` (`exception`, a `{"identifier", "name", "custom"}` kind, only for `WITH`). |
| `*.kind.identifier` | The `identifier` of a `license_kind`, in its canonical spelling for SPDX identifiers. |
| `*.kind.name` | The `full_name` of a `license_kind`. |
| `*.kind.custom` | `true` if the identifier is not on the SPDX license list. Absent otherwise. |
| `*.or_later` | `true` on a leaf that was followed by `+` in the `expression` (e.g., `GPL-3.0+`), meaning that the license or any later version applies. Absent otherwise. |
| `kind` | Deprecated, for tools that predate `expression`. The `{"identifier", "name"}` kind when the license is a single license without `+`, otherwise `null`. Tools should read `expression` instead. |
| `label` | The label of this `license` target. |
| `text` | The path of the license `text` file, relative to the execution root. Absent if no `text` was declared. |


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="license-name"></a>name |  A unique name for this target.   |  none |
| <a id="license-kind"></a>kind |  The kind of license this license is classified as. Exactly one of `kind` and `expression` is required.<br><br>This is typically a `license_kind` target. Targets for all SPDX licenses are predeclared in `@package_metadata//licenses/spdx`.   |  `None` |
| <a id="license-expression"></a>expression |  An [SPDX license expression](https://spdx.github.io/spdx-spec/v3.0.1/annexes/spdx-license-expressions/) (e.g., `MIT OR (Apache-2.0 AND BSD-3-Clause)`, `GPL-2.0-only WITH Classpath-exception-2.0`). Parsed at analysis time; malformed expressions fail the build. Standard SPDX identifiers are matched case-insensitively and reported with their canonical spelling, while the operators `AND`, `OR` and `WITH` are case-sensitive. A trailing `+` (e.g., `GPL-3.0+`) is kept as a flag on the license, not resolved to another identifier.   |  `None` |
| <a id="license-kinds"></a>kinds |  Additional `license_kind` targets referenced by `expression`, which are required for custom licenses (`LicenseRef-...`) and custom exceptions (`AdditionRef-...`, declared with `exception = True`). SPDX license and exception identifiers are resolved automatically from `@package_metadata//licenses/spdx`.   |  `None` |
| <a id="license-text"></a>text |  The [File](https://bazel.build/rules/lib/builtins/File) with the text of the license.<br><br>This is typically the `LICENSE` or `COPYING` file of the package. It is propagated to consumers so that it can be shipped alongside built artifacts.   |  `None` |
| <a id="license-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="license-visibility"></a>visibility |  The visibility of this target.   |  `None` |


