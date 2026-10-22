"""Declares rule `license`."""

load("//licenses/private:expression.bzl", "license_expression_ids", "parse_license_expression", "resolve_or_later")
load("//licenses/providers:license_expression_info.bzl", "LicenseExpressionInfo")
load("//licenses/providers:license_kind_info.bzl", "LicenseKindInfo")
load("//licenses/spdx:or_later.bzl", "OR_LATER")
load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `license` targets.
KIND = "build.bazel.attribute.license"

_SPDX_PACKAGE = "//licenses/spdx"

def _children(node, key):
    """Returns `(child, key)` pairs; keys are paths, as dicts aren't hashable."""
    if "operands" in node:
        return [(o, "{}.{}".format(key, i)) for i, o in enumerate(node["operands"])]
    if node.get("operator") == "WITH":
        return [(node["license"], key + ".w")]
    return []

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
            exception = kinds[node["exception"]]
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
                "exception": {"identifier": exception.identifier, "name": exception.name},
            }
        else:
            if node["or_later"] and node["id"] not in OR_LATER:
                fail("'{id}+' is not a valid license: SPDX defines no 'or later' identifier for '{id}'".format(id = node["id"]))
            kind = kinds[resolve_or_later(node, OR_LATER)]
            if kind.exception:
                fail("'{}' is a license exception and can only follow 'WITH'".format(kind.identifier))
            info = LicenseExpressionInfo(kind = kind)
            jsons[key] = {"kind": {"identifier": kind.identifier, "name": kind.name}}
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
        attribute["kind"] = {
            "identifier": kind.identifier,
            "name": kind.name,
        }
    else:
        tree = parse_license_expression(ctx.attr.expression)
        kinds = {}
        for k in ctx.attr.kinds:
            info = k[LicenseKindInfo]
            kinds[info.identifier] = info
        for kind_id in license_expression_ids(tree, OR_LATER):
            if kind_id not in kinds:
                fail("no license kind for '{}' in expression '{}'; pass it in `kinds`".format(kind_id, ctx.attr.expression))
        expression, attribute["expression"] = _build_graph(tree, kinds)
    providers = []
    files = []

    if ctx.attr.expression:
        providers.append(expression)

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
    provider of the `license` target.

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.license`, and its `attributes` file holds a JSON
    object, encoded as UTF-8 and written without insignificant whitespace (the
    example below is indented for readability):

    ```json
    {
        "kind": {
            "identifier": "GPL-3.0-or-later",
            "name": "GNU General Public License v3.0 or later"
        },
        "label": "@@//:license",
        "text": "COPYING"
    }
    ```

    | Field | Description |
    | :---- | :---------- |
    | `expression` | A graph of the parsed SPDX license expression, present instead of `kind` if `expression` was declared. A node is either a leaf `{"kind": {"identifier", "name"}}`  or `{"operator": "AND" or "OR" or "WITH", "operands": [...], "exception"}` (`exception`, a `{"identifier", "name"}` kind, only for `WITH`). |
    | `kind.identifier` | The `identifier` of the `kind` of this license. |
    | `kind.name` | The `full_name` of the `kind` of this license. |
    | `label` | The label of this `license` target. |
    | `text` | The path of the license `text` file, relative to the execution root. Absent if no `text` was declared. |

    Args:
      name: A unique name for this target.
      kind: The kind of license this license is classified as. Exactly one of
        `kind` and `expression` is required.

        This is typically a `license_kind` target. Targets for all SPDX
        licenses are predeclared in `@package_metadata//licenses/spdx`.
      expression: An [SPDX license expression](https://spdx.github.io/spdx-spec/v2.3.1/SPDX-license-expressions/)
        (e.g., `MIT OR (Apache-2.0 AND BSD-3-Clause)`, `GPL-2.0-only WITH
        Classpath-exception-2.0`). Parsed at analysis time; malformed
        expressions fail the build. `<id>+` is resolved to its SPDX identifier
        (`GPL-3.0+` becomes `GPL-3.0-or-later`); it is invalid if SPDX defines
        no such identifier (e.g., `Apache-2.0+`).
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
        extra = {k: True for k in (kinds or [])}
        predeclared = [
            Label(_SPDX_PACKAGE + ":" + i)
            for i in license_expression_ids(parse_license_expression(expression), OR_LATER)
            if not i.startswith(("LicenseRef-", "AdditionRef-")) and ":" not in i
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
