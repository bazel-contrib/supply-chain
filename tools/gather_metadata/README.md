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

## Declaring toolchain involvement on rules

Rulesets can declare how their toolchains participate in SBOMs directly in the
rule definition by adding the private `_supply_chain_toolchain_involvement`
label-keyed string-dictionary attribute:

```starlark
my_binary = rule(
    implementation = _impl,
    toolchains = ["//:java_type", "//:node_type"],
    attrs = {
        "srcs": attr.label_list(),
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("//:java_type"): "build",
            Label("//:node_type"): "embedded",
        }),
    },
)

```

Each key is a toolchain-type label and each value is `"runtime"`, `"embedded"`,
or `"build"`. Since the keys are real label dependencies, Bazel resolves their
repository mapping and reports misspelled or unavailable labels at loading
time. The aspect matches the declaration to an edge by toolchain type and
validates that each key names a toolchain declared by the rule, reporting
invalid keys or values with the rule kind and target label. A rule without the attribute retains the existing heuristics:
exec-configured toolchains are `"build"` and other toolchains are `"runtime"`.

This design keeps declarations next to rules, avoids a shared generated
registry, and limits invalidation to rule definitions that declare the
attribute. The trade-off is that root modules can no longer override a
ruleset's declaration; they must patch the rule definition instead.

The `attr.label_keyed_string_dict` contract is supported by the repository's
Bazel 7.x, 8.x, and 9.x CI lines. The aspect excludes this private attribute
from dependency traversal so its label keys do not become SBOM edges.

### Alternatives considered

The [rules_runfiles_group](https://github.com/bazel-contrib/rules_runfiles_group)
pattern uses a shared `_aspect_rule_hints` label-list with provider-carrying
hint targets, avoiding a separate attribute for every concern. That approach
requires instantiating a target per rule, and merging a hint attr with `|`
would clobber any existing `_aspect_rule_hints` definition. This experiment
therefore uses plain-data private string-dict attributes, the documented
`_supply_chain_toolchain_involvement` contract.

## Provider-declared dependency relationships

Rulesets can classify dependency edges independently of toolchain involvement
by adding the private `_supply_chain_relationships` string-dictionary
attribute to the rule definition:

```starlark
image_layer = rule(
    implementation = _impl,
    attrs = {
        "srcs": attr.label_list(),
        "_supply_chain_relationships": attr.string_dict(default = {
            "srcs": "contains",
        }),
    },
)
```

Keys are attribute names declared by that rule. Values must be `depends_on`,
`contains`, `static_link`, `dynamic_link`, `test_dependency`, or
`dev_dependency`. The aspect validates the attribute type, keys, and values at
analysis time and reports the rule kind, target label, and attribute name for
invalid declarations. Undeclared attributes and toolchain edges default to
`depends_on`. Since each rule carries its own declaration, same-named rules in
different modules can naturally use different relationships. As with
toolchain involvement, a root module cannot override a declaration without
patching the owning rule; there is no typo safety at definition time.

The relationship attribute only changes classification; the serialized graph
and Go generators retain the same edge fields and vocabulary. When graph paths
collapse intermediate Bazel targets, the first non-default relationship
encountered from the root is retained. Toolchain `involvement` remains
orthogonal.

| Relationship | SPDX 2.3 | SPDX 3.0.1 | CycloneDX |
| --- | --- | --- | --- |
| `depends_on` | parent `DEPENDS_ON` child | `dependsOn` | dependency edge |
| `contains` | parent `CONTAINS` child | `contains` | parent property `bazel:relationship:contains` |
| `static_link` | parent `STATIC_LINK` child | `hasStaticLink` | dependency edge and parent property `bazel:relationship:static_link` |
| `dynamic_link` | parent `DYNAMIC_LINK` child | `hasDynamicLink` | dependency edge and parent property `bazel:relationship:dynamic_link` |
| `test_dependency` | child `TEST_DEPENDENCY_OF` parent | `dependsOn`, lifecycle scope `test` | dependency edge and parent property `bazel:relationship:test_dependency` |
| `dev_dependency` | child `DEV_DEPENDENCY_OF` parent | `dependsOn`, lifecycle scope `development` | dependency edge and parent property `bazel:relationship:dev_dependency` |

CycloneDX has no corresponding relationship vocabulary. `contains` edges are
represented by a `bazel:relationship:contains` property on the parent
component, with the bom-ref of the child as its value; they are not emitted as
dependency edges, and no BOM-level compositions are generated. Embedded
involvement is treated as `contains`, unless build involvement takes
precedence. Other non-default relationship types remain dependency edges and
are represented by sorted, deduplicated properties on the parent. Default
`depends_on` edges need no extra property. Components remain flat. Containment
is never cut and does not appear as a dependency edge, but it is followed from
the entry components when breaking dependency cycles, so that the dependencies
of contained components, directly or transitively, are preserved.
