# Examples for supply_chain_tools

This set of files provides an example of how license rules can be used.

Terminology
-   Organization: A company, organization or other entity that wants to use
    license rules to enforce their particular compliance needs. These examples
    use the work organization throughout.
-   SCM: source code management system. These examples assume that
    an organization has a SCM that can enforce ownership restrictions on
    specific folder trees. Targets are divided into BUILD files that are
    reviewed by engineers vs. those that are reviewed by an organizations
    compliance team.

## Overview

This directory is a standalone Bazel module. Its `MODULE.bazel` uses local
overrides so the examples exercise the metadata rules and SBOM tools from this
checkout.

The SBOM example reports on `//lib:lib_with_third_party_deps`, which depends on
`//vendor/acme`. The library uses the example module's package metadata; ACME
declares its own package URL, version, and license metadata.

## Generate an SBOM

With Bazelisk installed, run from this directory:

```sh
bazel build \
    //sample_reports:library_cyclonedx \
    //sample_reports:library_spdx
```

The first build downloads Bazel dependencies and a Go toolchain, so it needs
network access. The generated JSON files are:

- `bazel-bin/sample_reports/library_cyclonedx.json`
- `bazel-bin/sample_reports/library_spdx.json`

Both reports describe the example module and its ACME dependency
(`pkg:github/acmecorp/coyote@3.1.4`). These are demonstration package identities,
not an inventory of the tools used to build the reports.

The targets in `sample_reports/BUILD` wire together three stages:

1. `sbom` traverses the target's dependencies, gathers package metadata, and
   classifies the root component and its dependencies.
2. `cyclonedx` exports that graph and classification as CycloneDX JSON.
3. `spdx` exports the same inputs as SPDX JSON.

The intermediate files are also available under `bazel-bin/sample_reports/`:
`library_sbom.graph.json` and `library_sbom.sbom-classifications.json`.
Metadata file paths in the graph are relative to Bazel's execution root.

To report on another target, add an `sbom` rule pointing to it and pass that rule
to a `cyclonedx` or `spdx` rule. Attach `package_metadata` to the target and its
dependencies with `package(default_package_metadata = [...])`. The `sbom` rule
requires root metadata by default; `require_root_metadata = False` permits
reports without a root component.

The exporters also support `format = "xml"` for CycloneDX and
`format = "yaml"` or `format = "tag-value"` for SPDX.
