# License coverage and policy

The Go tooling module contains reusable building blocks for organizations that
want to enforce license metadata without coupling policy to a particular
repository or language:

- `//lib/supplychain-go/coverage` turns the graph emitted by
  `//tools/gather_metadata` into a deterministic coverage report.
- `//lib/supplychain-go/license_resolver` classifies license files in a fetched
  source directory using the upstream `licenseclassifier` corpus. The input is
  a source directory, so callers can use it for Go modules, Cargo crates,
  Maven artifacts, or another package ecosystem.
- `//lib/supplychain-go/policy` evaluates the coverage report against an
  organization-owned JSON policy. SPDX `AND`, `OR`, and `WITH` expressions are
  evaluated structurally, and exceptions can be scoped to one exact artifact
  root with an expiry, rationale, and organization-defined scope.
- `//tools/sbom:supply_chain.bzl` exposes a single front door that composes the
  coverage, policy, SPDX, and CycloneDX steps as ordinary Bazel targets.

The repository intentionally does not ship a default allowlist. Each
organization owns its legal and compliance decisions and should keep that
policy data with its own review controls.

## Resolver input

`//lib/supplychain-go/cmd/license_resolver` accepts:

```json
{
  "packages": [
    {
      "purl": "pkg:golang/example.com/library@v1.2.3",
      "root": "/path/to/fetched/source",
      "source_metadata_label": "@@go_deps+example.com/library//:package_metadata"
    }
  ]
}
```

It emits one result per package, including the conservative SPDX expression and
the files that supplied the evidence. It never writes a derived license
catalog; callers can use the output directly as an input to coverage.

## End-to-end Bazel usage

Add the tools module as a dependency of the workspace that owns the artifact,
then declare one front-door target:

The example assumes `:license_resolutions` is a generated file target from a
language-specific resolver. Omit `resolutions` when package metadata already
contains all required licenses.

```starlark
load("@supply_chain_tools//sbom:supply_chain.bzl", "supply_chain")

supply_chain(
    name = "app_license_compliance",
    target = "//app:binary",
    policy = ":license-policy.json",
    # Optional: a transient output from a language-specific resolver. Omit
    # this when package metadata already contains every required license.
    resolutions = ":license_resolutions",
    sbom_formats = ["spdx-json", "cyclonedx-json"],
)
```

The policy file is owned by the consuming organization. A minimal example is:

```json
{
  "schema_version": 1,
  "allowed_license_identifiers": ["Apache-2.0", "MIT"],
  "allowed_license_exceptions": ["LLVM-exception"],
  "allowed_missing_license_purls": [],
  "target_exceptions": []
}
```

Build the policy target in CI:

```shell
bazel build //app:app_license_compliance
```

The action succeeds only when every package has license metadata or an explicit
missing-license allowance, and every reported SPDX expression is acceptable.
The evaluation JSON, coverage report, and requested SBOM documents are
available as the target's outputs on success. To inspect coverage without
enforcing policy, omit `policy` and build the same front-door target.

The macro also creates predictable private intermediate targets named
`<name>_sbom`, `<name>_license_coverage`, and (when `policy` is set)
`<name>_license_policy`. They are useful for debugging, but consumers should
normally depend only on the single `<name>` target.

### Go/Bazel source discovery

Go dependencies fetched by Gazelle can use the built-in adapter. It maps
external `pkg:golang` metadata labels to `external/<repository>` under the
Bazel output base and feeds those roots to the same generic classifier:

```shell
bazel run @supply-chain-go//cmd/go_license_resolver -- \
  --graph bazel-bin/app/app_license_compliance_sbom.graph.json \
  --output /tmp/go-license-resolutions.json \
  --workspace "$PWD"
```

Pass that JSON as the `resolutions` input to `supply_chain`. Other ecosystems
can implement the same small `{purl, root}` adapter contract without changing
the classifier, coverage report, policy evaluator, or SBOM emitters.

## CLI-only usage

The resolver can be run independently when a language integration has mapped a
package to its fetched source directory:

```shell
bazel run @supply-chain-go//cmd/license_resolver -- \
  --input "$PWD/license-resolver-input.json" \
  --output "$PWD/license-resolutions.json"
```

Feed that output to `license_coverage`, or evaluate an existing report
directly:

```shell
bazel run @supply-chain-go//cmd/license_policy -- \
  --report "$PWD/license-coverage.json" \
  --allowlist "$PWD/license-policy.json"
```

The policy command prints a JSON evaluation and exits non-zero on a policy
failure, which makes it suitable for CI and pre-submit checks. The source-tree
resolver is intentionally language-neutral. Go/Bazel has a built-in adapter;
Cargo, Maven, and other integrations are responsible only for producing the
same `{purl, root}` input until their source-root conventions are standardized.

## Policy input

`//lib/supplychain-go/cmd/license_policy` accepts an organization-owned policy
file and coverage report. It emits JSON and exits non-zero when metadata is
missing, a license expression is unacceptable, or a scoped exception is
expired or does not match the exact artifact root.
