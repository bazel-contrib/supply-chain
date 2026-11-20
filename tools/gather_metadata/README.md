# Metadata Gathering System

Bazel rules and aspects to walk dependency trees and gather metadata providers (licenses, package information, etc.).

This set of rules produces a graph of nodes (targets) and their edges starting from a root target. This can be automatically serialized to JSON and looks like the example below:

```json
{
  "schema_version": "1.0",
  "root_target": "//some:target",
  "nodes": [
    {
      "label": "//some:target",
      "metadata_file": "path/to/target.package-metadata.json"
    }
  ],
  "edges": [
    {
      "from": "//some:target",
      "to": "//other:dep",
      "type": "depends_on",
      "involvement": "embedded",
      "relationship": "contains"
    }
  ]
}
```

Edges omit `involvement` for runtime dependencies (the default). Non-runtime
edges use `"build"` for dependencies used only while producing the output, or
`"embedded"` when dependency code is incorporated into the output. The field
is optional so existing graph consumers and runtime-only graph output remain
unchanged.

## Provider-declared toolchain involvement

Toolchain providers can declare how their toolchains participate in SBOMs from
their own `MODULE.bazel`. Add a dependency on `supply_chain_tools`, then declare
tags through its `toolchain_involvement` module extension:

```starlark
bazel_dep(name = "supply_chain_tools", version = "...")

toolchain_involvement = use_extension(
    "@supply_chain_tools//gather_metadata:extensions.bzl",
    "toolchain_involvement",
)
toolchain_involvement.declare(
    name = "java-compiler",
    toolchain_type = "@bazel_tools//tools/jdk:toolchain_type",
    rule_kind = "java_binary",
    involvement = "build",
)
```

`declare` accepts a required `toolchain_type` label and `involvement`
(`"runtime"`, `"embedded"`, or `"build"`). `rule_kind` is optional and
defaults to `"*"`, meaning any rule kind; use a specific rule kind when the
toolchain's role depends on how it is consumed. `name` is an optional
human-readable identifier used in conflict diagnostics. The label is resolved
in the declaring module's repository mapping, so providers can use their
apparent repository names in `MODULE.bazel`.

The extension merges declarations from every module that uses it and generates
`@supply_chain_toolchain_involvements//:toolchain_involvements`. Precedence is:

1. A declaration in the root module wins for the same `(rule_kind,
   toolchain_type)` key.
2. If the root has no declaration, identical provider declarations coalesce;
   conflicting declarations from non-root modules fail with both module names
   and values in the error.

The demo in `examples/complex_app` illustrates real toolchain labels used
with its pinned modules (confirmed from their resolved toolchain targets):

```starlark
toolchain_involvement.declare(
    name = "go",
    toolchain_type = "@io_bazel_rules_go//go:toolchain",
    rule_kind = "go_binary",
    involvement = "build",
)
toolchain_involvement.declare(
    name = "llvm-cgo",
    toolchain_type = "@bazel_tools//tools/cpp:toolchain_type",
    rule_kind = "go_binary",
    involvement = "embedded",
)
toolchain_involvement.declare(
    name = "java-compiler",
    toolchain_type = "@bazel_tools//tools/jdk:toolchain_type",
    rule_kind = "java_binary",
    involvement = "build",
)
toolchain_involvement.declare(
    name = "python-image",
    toolchain_type = "@bazel_tools//tools/python:toolchain_type",
    rule_kind = "py_image_layer",
    involvement = "embedded",
)
toolchain_involvement.declare(
    name = "node-host-tool",
    toolchain_type = "@rules_nodejs//nodejs:toolchain_type",
    involvement = "build",
)
```

The Java runtime type is
`@bazel_tools//tools/jdk:runtime_toolchain_type`; it defaults to runtime in
target configuration, while the Java compiler type above is build-involved.
`rules_nodejs` similarly distinguishes its host `toolchain_type` from
`runtime_toolchain_type`. The latter is runtime-involved for target-configured
programs, but can be build-involved for exec-configured tools. Since registry
keys intentionally do not include configuration, avoid a blanket runtime
override for rules used in both configurations and let the aspect classify
those edges from the configuration. The demo declares root-module tags on
behalf of its pinned upstream providers; each provider can move the equivalent
declaration into its own `MODULE.bazel`.

`single_version_override` is a root-module-only directive: an override written
by a dependency is ignored when that dependency is consumed. The demo uses root
tags as a provider-side stand-in for toolchain involvement declarations.

## Provider-declared dependency relationships

Modules can classify rule-attribute edges independently of toolchain
involvement. A provider declares its rule relationships in its own
`MODULE.bazel` through the `relationships` extension:

```starlark
relationships = use_extension(
    "@supply_chain_tools//gather_metadata:extensions.bzl",
    "relationships",
)
relationships.declare(
    name = "image-layer-files",
    rule_kind = "image_layer",
    attribute = "srcs",
    relationship = "contains",
)
```

For provider declarations to be module-specific, patch each declared rule with
a private string attribute in its `attrs` map:

```starlark
"_supply_chain_module": attr.string(default = Label("//:x").repo_name),
```

The aspect reads this marker from `ctx.rule.attr` and normalizes a canonical
Bazel repository name (for example, `rules_img+`) to its Bzlmod module name.
Provider-scoped tags do not match unmarked rules; a root fallback may still
classify them. This keeps identically named rules from different modules
separate. The `examples/complex_app` patches show declarations added to
the `rules_img` and `rules_pkg` modules and markers added to their actual rule
definitions.

The vocabulary is `depends_on` (the default and omitted from serialized
edges), `contains`, `static_link`, `dynamic_link`, `test_dependency`, and
`dev_dependency`. `name` is an optional diagnostic label. Rule kinds may be
`*` to match all kinds; for a given
attribute, an exact rule-kind declaration takes precedence over the wildcard.
The extension generates `@supply_chain_relationships//:relationships` even
when no declarations exist. Unknown vocabulary values fail during extension
evaluation.

The registry is keyed by `(rule_module, rule_kind, attribute,
requires_attributes)`. Provider declarations are automatically scoped to the
declaring module and match only rules whose private marker identifies that
module. A root declaration may set `rule_module` to target a patched module
explicitly. Without it, the declaration has no module scope and acts as a
kind-based fallback for unmarked rules or marked rules without a more-specific
module declaration. Lookup prefers an exact module marker over that fallback,
then an exact rule kind over `*`, then the declaration with the most matching
`requires_attributes`.

`requires_attributes` is optional and is primarily useful for root fallbacks
when two unmarked rule definitions share a kind and attribute. Every listed
name must appear in `dir(ctx.rule.attr)`; Bazel includes declared attributes
even when they are unset. Equal-specificity declarations that match one rule
and disagree on the relationship fail during analysis with both declaring
module names and values.

For the same full registry key, identical provider declarations coalesce. A
root declaration wins over a provider declaration; conflicting declarations
from non-root modules fail during extension evaluation with module names and
values. A root declaration with `rule_module` can therefore override a
provider's declaration for that module. There are no built-in relationship
defaults: an unmatched edge remains `depends_on`.

When graph paths collapse intermediate Bazel targets, the first non-default
relationship encountered from the root is retained. If multiple edges map to
the same component pair, relationship specificity is `contains`, `static_link`,
`dynamic_link`, `test_dependency`, `dev_dependency`, then `depends_on`.
Toolchain `involvement` remains orthogonal; build/embedded involvement keeps
its existing output behavior and takes precedence over the attribute
relationship when rendering.

| Relationship | SPDX 2.3 | SPDX 3.0.1 | CycloneDX |
| --- | --- | --- | --- |
| `depends_on` | parent `DEPENDS_ON` child | `dependsOn` | dependency edge |
| `contains` | parent `CONTAINS` child | `contains` | BOM composition with `aggregate: "unknown"` and child bom-ref in `assemblies` |
| `static_link` | parent `STATIC_LINK` child | `hasStaticLink` | dependency edge and parent property `bazel:relationship:static_link` |
| `dynamic_link` | parent `DYNAMIC_LINK` child | `hasDynamicLink` | dependency edge and parent property `bazel:relationship:dynamic_link` |
| `test_dependency` | child `TEST_DEPENDENCY_OF` parent | `dependsOn`, lifecycle scope `test` | dependency edge and parent property `bazel:relationship:test_dependency` |
| `dev_dependency` | child `DEV_DEPENDENCY_OF` parent | `dependsOn`, lifecycle scope `development` | dependency edge and parent property `bazel:relationship:dev_dependency` |

CycloneDX has no corresponding relationship vocabulary. `contains` edges are
represented by BOM-level compositions, one per parent bom-ref, with
`aggregate: "unknown"` and sorted child bom-refs in `assemblies`; they are not
emitted as dependency edges or `bazel:relationship:contains` properties.
Embedded involvement also maps to compositions, unless build involvement
takes precedence. Other non-default relationship types remain dependency
edges and are represented by sorted, deduplicated properties on the parent.
Default `depends_on` edges need no extra property. Components remain flat, and
composition references do not participate in dependency-cycle removal.
