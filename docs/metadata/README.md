# Module `@package_metadata`

General-purpose rules for injecting supply-chain metadata into Bazel projects (e.g., for generating [SBOM](https://www.ntia.gov/page/software-bill-materials)s for shipped software artifacts).


## Stability

This is a fundamental module of the Bazel ecosystem that most, if not all, other Bazel modules depend on. Stability is therefore very important to us: once a symbol is declared stable, we will not make breaking changes to it.

Stability is tracked per symbol, not for the module as a whole. Every symbol listed under [API Documentation](#api-documentation) is covered by the guarantee unless an exception is called out explicitly — either in this document (see [attributes](#as-module-author) and [the consumer-side API](#as-an-organization)) or in the symbol's own API documentation.


## Concepts

Requirements for software supply-chain security measures vary widely depending on the organization building the product or the jurisdiction(s) the software is shipped to. They are also subject to change over time as laws evolve over time and companies find themselves in need to comply to these new requirements from governments around the world. Hence, the rules to inject supply-chain metadata needs to be very customizable.

The core of this module is built around `package`s and `attribute`s.

  - `package` is used to identify (third-party) software and track its origin (e.g., `npm` module, `maven` artifact, or Rust `crate` its downloaded from). We use [PURL](https://github.com/package-url/purl-spec)s for this.
  - `attribute`s are used to declare metadata attached to `package`s (e.g., the License, or who the maintainers are). They are identified by `kind` to distinguish between different types of `attribute`s.
    - This provides the primary extension point for organizations to inject the metadata they need.
    - While we encourage organizations to adopt "well-known" `attribute`s provided in this module whenever possible, custom `attribute`s are also expected.


## Usage

The rules and providers in this module have two primary audiences:

  - Authors of modules in the [Bazel Central Registry](https://registry.bazel.build) (or private registries) that want to annotate their module/packages/targets, and
  - Organizations that want to consume annotations for compliance checks or for producing provenance information for artifacts.

### As module author

If you are a module author and want to annotate your module, you will need :

  - Choose the appropriate package-url(PURL) type from the [standard list](https://github.com/package-url/purl-spec/tree/main/types-doc) related to the module technology if available. If no package type is already defined, use the `generic` type and provide the appropriate qualifiers to unambiguously identify the package.

  - Add a dependency on `package_metadata` to your `MODULE.bazel` file.

    ```starlark
    bazel_dep(name = "package_metadata", version = "<check releases>")
    ```

  - Create `package_metadata` target(s) for declaring metadata.

    This target is typically in the top-level `BUILD.bazel` file.

    > Modules typically need only a single `package_metadata` target. However, multiple targets can be required in some cases (e.g., when some targets are licensed under a different license).

    ```starlark
    load("@package_metadata//purl:purl.bzl", "purl")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            # ...
        ],
        purl = purl.bazel(module_name(), module_version()),
        visibility = ["//visibility:public"],
    )
    ```

  - (optional) Add `attributes` to your `package_metadata` target(s).

    `package_metadata` itself only provides information about the identity of a module or package and where it was retrieved from. Additional metadata is provided as `attributes` to the `package_metadata` target (e.g., the license the packages are under, ...).

    The attributes provided by this module are:

      - [license](./licenses/rules/license.md#license) — the license the package is available under.
      - [copyright_notice](./attributes/copyright_notice.md#copyright_notice) — the copyright notices of the package.

    Custom attributes are declared by writing a rule that provides [PackageAttributeInfo](./providers/package_attribute_info.md#PackageAttributeInfo). [`copyright_notice`](../../metadata/attributes/copyright_notice.bzl) is a minimal example to copy from.

    > **IMPORTANT**: The extension point for custom attributes, and the attributes other than `license`, are still under development. Please avoid relying on them in modules published to a public registry for now.

  - Annotate all targets with `package_metadata`.

    This step is required for consumers to access the declared metadata.

    There are three options to annotate targets:

      - Module Level: Add `default_package_metadata` to `REPO.bazel`

        It requires changing a single file only.

        ```starlark
        repo(default_package_metadata = ["//:package_metadata"])
        ```

        This provides a simple way to annotate all targets in a module, while preserving the ability to annotate individual packages or targets with different metadata using the methods below.

      - Package level: Add `default_package_metadata` to all packages

        Similar approach to adding `default_package_metadata` to `REPO.bazel`, but on a per package level.

        ```starlark
        package(default_package_metadata = ["//:package_metadata"])
        ```

        This provides a simple way to annotate all targets in a package, while preserving the ability to annotate individual targets in the package with different metadata using the method below.

      - Target level: Add `package_metadata` to all targets individually:

        ```starlark
        foo_library(
            name = "hello",
            package_metadata = [
                "//:package_metadata",
            ],
            # ...
        )
        ```

        While this allows very fine grained control over the metadata of a target, it's also very tedious to modify all targets in a module. This method should therefore be reserved for targets with different metadata.


  - Publish your module.

### As an organization

Organizations consume the metadata that module authors declare — to check compliance, or to produce provenance information for the artifacts they ship.

The rules and providers in this module only *declare* metadata. Collecting it across a build graph and rendering it into a report (e.g., an SBOM) is the job of the [`@supply_chain_tools`](../../tools) module.

One consumer-side rule is available already, in a separate module: to attach metadata to a dependency that does not declare any itself (e.g., a third-party module that has not adopted `package_metadata` yet), use [package_metadata_override](../metadata-extensions/rules/package_metadata_override.md#package_metadata_override) from [@package_metadata_extensions](../metadata-extensions). The providers it builds on, [PackageMetadataOverrideInfo](./providers/package_metadata_override_info.md#PackageMetadataOverrideInfo) and [PackageMetadataToolchainInfo](./providers/package_metadata_toolchain_info.md#PackageMetadataToolchainInfo), are part of `@package_metadata` itself.

> **IMPORTANT**: The consumer-side API is under active development and changes more frequently than the declaration-side API described above. We will document it here after stabilizing it.


## API Documentation

Where a package has a `defs.bzl` (e.g., `@package_metadata//:defs.bzl`), it re-exports the public symbols of that package. The pages below document each symbol under its canonical per-file path; both load paths are supported.

### Generic

  - [@package_metadata//:defs.bzl](./defs.md)

#### Providers

  - [@package_metadata//providers:package_attribute_info.bzl](./providers/package_attribute_info.md)
  - [@package_metadata//providers:package_metadata_info.bzl](./providers/package_metadata_info.md)
  - [@package_metadata//providers:package_metadata_override_info.bzl](./providers/package_metadata_override_info.md)
  - [@package_metadata//providers:package_metadata_toolchain_info.bzl](./providers/package_metadata_toolchain_info.md)

#### Rules

  - [@package_metadata//rules:package_metadata.bzl](./rules/package_metadata.md)

#### Utils

  - [@package_metadata//common:common.bzl](./common/common.md)
  - [@package_metadata//purl:purl.bzl](./purl/purl.md)


### Attributes

  - [@package_metadata//attributes:copyright_notice.bzl](./attributes/copyright_notice.md)


### Licenses

  - [@package_metadata//licenses:defs.bzl](./licenses/defs.md)

#### Providers

  - [@package_metadata//licenses/providers:license_kind_info.bzl](./licenses/providers/license_kind_info.md)

#### Rules

  - [@package_metadata//licenses/rules:license.bzl](./licenses/rules/license.md)
  - [@package_metadata//licenses/rules:license_kind.bzl](./licenses/rules/license_kind.md)
