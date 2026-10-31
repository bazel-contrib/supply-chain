"""Declares rule `license`."""

load("//licenses/private:expression.bzl", "canonical_id", "license_expression_ids", "parse_license_expression")
load("//licenses/providers:license_expression_info.bzl", "LicenseExpressionInfo")
load("//licenses/providers:license_kind_info.bzl", "LicenseKindInfo")
load("//licenses/spdx:ids.bzl", "SPDX_IDS")
load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `license` targets.
KIND = "build.bazel.attribute.license"

_SPDX_PACKAGE = "//licenses/spdx"

def _kind_json(kind):
    """Returns the JSON of a `LicenseKindInfo`, flagging non-SPDX identifiers as custom."""
    result = {"identifier": kind.identifier, "name": kind.name}
    if SPDX_IDS.get(kind.identifier.lower()) != kind.identifier:
        result["custom"] = True
    return result

def _children(node, key):
    """Returns `(child, key)` pairs; keys are paths, as dicts aren't hashable."""
    if "operands" in node:
        return [(o, "{}.{}".format(key, i)) for i, o in enumerate(node["operands"])]
    if node.get("operator") == "WITH":
        return [(node["license"], key + ".w")]
    return []

def _check_not_exception(kind):
    if kind.exception:
        fail("'{}' is a license exception and can only follow 'WITH'".format(kind.identifier))

def _build_graph(tree, kinds):
    """Converts a parsed expression into `LicenseExpressionInfo`s.

    Starlark forbids recursion, so this builds bottom-up over an explicit
    traversal.
    """
    order = []
    stack = [(tree, "r")]
    for _ in range(100000):
        if not stack:
            break
        item = stack.pop()
        order.append(item)
        stack.extend(_children(item[0], item[1]))

    built = {}
    jsons = {}
    for node, key in reversed(order):
        child_keys = [k for _, k in _children(node, key)]
        children = [built[k] for k in child_keys]
        child_jsons = [jsons[k] for k in child_keys]
        if "operands" in node:
            info = LicenseExpressionInfo(operator = node["operator"], operands = children)
            jsons[key] = {"operator": node["operator"], "operands": child_jsons}
        elif node.get("operator") == "WITH":
            exception = kinds[canonical_id(node["exception"], SPDX_IDS)]
            if not exception.exception:
                fail("'{}' is not a license exception and cannot follow 'WITH'; declare its `license_kind` with `exception = True`".format(exception.identifier))
            info = LicenseExpressionInfo(
                operator = "WITH",
                operands = children,
                exception = exception,
            )
            jsons[key] = {
                "operator": "WITH",
                "operands": child_jsons,
                "exception": _kind_json(exception),
            }
        else:
            kind = kinds[canonical_id(node["id"], SPDX_IDS)]
            _check_not_exception(kind)
            info = LicenseExpressionInfo(kind = kind, or_later = node["or_later"])
            jsons[key] = {"kind": _kind_json(kind)}
            if node["or_later"]:
                jsons[key]["or_later"] = True
        built[key] = info
    return built["r"], jsons["r"]

def _license_impl(ctx):
    if bool(ctx.attr.kind) == bool(ctx.attr.expression):
        fail("exactly one of `kind` and `expression` must be set")

    attribute = {
        "label": str(ctx.label),
    }
    if ctx.attr.kind:
        kind = ctx.attr.kind[LicenseKindInfo]
        _check_not_exception(kind)
        attribute["expression"] = {"kind": _kind_json(kind)}
        expression = LicenseExpressionInfo(kind = kind)
    else:
        tree = parse_license_expression(ctx.attr.expression)
        kinds = {}
        for k in ctx.attr.kinds:
            info = k[LicenseKindInfo]
            kinds[info.identifier] = info
        for kind_id in license_expression_ids(tree, SPDX_IDS):
            if kind_id not in kinds:
                fail("no license kind for '{}' in expression '{}'; pass it in `kinds`".format(kind_id, ctx.attr.expression))
        expression, attribute["expression"] = _build_graph(tree, kinds)

    # Deprecated: `kind` is kept for tools that predate `expression`. It is
    # only set for a single license without `+`, which they can represent.
    root = attribute["expression"]
    attribute["kind"] = root["kind"] if "kind" in root and not root.get("or_later") else None
    providers = [expression]
    files = []

    if ctx.attr.text:
        attribute["text"] = ctx.file.text.path
        files.append(ctx.attr.text[DefaultInfo].files)

    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        output = output,
        content = json.encode(attribute),
    )

    return providers + [
        DefaultInfo(
            files = depset(
                direct = [
                    output,
                ],
            ),
        ),
        PackageAttributeInfo(
            kind = KIND,
            attributes = output,
            files = files,
        ),
    ]

# Documentation lives on the public `license` macro below: Stardoc renders the
# macro, not this private rule.
_license = rule(
    implementation = _license_impl,
    attrs = {
        "expression": attr.string(),
        "kind": attr.label(
            providers = [
                LicenseKindInfo,
            ],
        ),
        "kinds": attr.label_list(
            providers = [
                LicenseKindInfo,
            ],
        ),
        "text": attr.label(
            mandatory = False,
            allow_single_file = True,
        ),
    },
)

def license(
        # Disallow unnamed attributes.
        *,
        # `_license` attributes.
        name,
        kind = None,
        expression = None,
        kinds = None,
        text = None,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring the license of a package or target.

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

    Args:
      name: A unique name for this target.
      kind: The kind of license this license is classified as. Exactly one of
        `kind` and `expression` is required.

        This is typically a `license_kind` target. Targets for all SPDX
        licenses are predeclared in `@package_metadata//licenses/spdx`.
      expression: An [SPDX license expression](https://spdx.github.io/spdx-spec/v3.0.1/annexes/spdx-license-expressions/)
        (e.g., `MIT OR (Apache-2.0 AND BSD-3-Clause)`, `GPL-2.0-only WITH
        Classpath-exception-2.0`). Parsed at analysis time; malformed
        expressions fail the build. Standard SPDX identifiers are matched
        case-insensitively and reported with their canonical spelling, while
        the operators `AND`, `OR` and `WITH` are case-sensitive. A trailing `+`
        (e.g., `GPL-3.0+`) is kept as a flag on the license, not resolved to
        another identifier.
      kinds: Additional `license_kind` targets referenced by `expression`, which
        are required for custom licenses (`LicenseRef-...`) and custom
        exceptions (`AdditionRef-...`, declared with `exception = True`). SPDX
        license and exception identifiers are resolved automatically from
        `@package_metadata//licenses/spdx`.
      text: The [File](https://bazel.build/rules/lib/builtins/File) with the
        text of the license.

        This is typically the `LICENSE` or `COPYING` file of the package. It is
        propagated to consumers so that it can be shipped alongside built
        artifacts.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    if expression:
        # Every SPDX identifier resolves to its predeclared `license_kind`;
        # `kinds` supplies the rest (`LicenseRef-` and `AdditionRef-` identifiers).
        # Labels given as strings are resolved before being compared.
        extra = {native.package_relative_label(k): True for k in (kinds or [])}
        predeclared = [
            Label(_SPDX_PACKAGE + ":" + i)
            for i in license_expression_ids(parse_license_expression(expression), SPDX_IDS)
            if SPDX_IDS.get(i.lower()) == i
        ]
        kinds = list(kinds or []) + [k for k in predeclared if k not in extra]

    _license(
        # `_license` attributes.
        name = name,
        kind = kind,
        expression = expression,
        kinds = kinds,
        text = text,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        applicable_licenses = [],
    )
