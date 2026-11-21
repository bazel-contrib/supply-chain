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
      "involvement": "embedded"
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
time. The aspect normalizes canonical toolchain labels when matching the
declaration to an edge and validates that each key names a toolchain declared
by the rule, reporting invalid keys or values with the rule kind and target
label. A rule without the attribute retains the existing heuristics:
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
