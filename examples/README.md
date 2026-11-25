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
    //sample_reports:library_spdx \
    //sample_reports:library_spdx_3_0_1
```

The first build downloads Bazel dependencies and a Go toolchain, so it needs
network access. The generated JSON files are:

- `bazel-bin/sample_reports/library_cyclonedx.json`
- `bazel-bin/sample_reports/library_spdx.json` (SPDX 2.3)
- `bazel-bin/sample_reports/library_spdx_3_0_1.json` (SPDX 3.0.1)

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

The `cyclonedx` exporter also supports `format = "xml"`. The `spdx` exporter
writes SPDX 2.3 as `format = "json"` (the default), `"yaml"` or `"tag-value"`,
and SPDX 3.0.1 as `format = "json-ld"`.

Both exporters record when the report was generated (`metadata.timestamp` in
CycloneDX, `creationInfo.created` in SPDX): the build time with `--stamp`, and a
fixed value otherwise, so that reports stay reproducible. To make a report
unique to a build, give `cyclonedx` a `serial_number` (a file holding a
`urn:uuid:` value) or `spdx` a `document_namespace` (a file holding an absolute
URI without a fragment). Without them a CycloneDX report has no serial number,
and the SPDX namespace is derived from the purl of the subject. They are files
so that a build can produce them, for example from its workspace status with
`expand_template`.
