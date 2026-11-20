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

The demo in `examples/toolchain_demo` illustrates real toolchain labels used
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
by a dependency is ignored when that dependency is consumed. Bazel's
documentation says a patch to a module's `MODULE.bazel` is effective only when
the patch is provided by the root module; the demo does not patch upstream
module declarations and instead uses root tags as the provider-side stand-in.
