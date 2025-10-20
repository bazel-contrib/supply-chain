# Module `@package_metadata`

General-purpose rules for injecting supply-chain metadata into Bazel projects (e.g., for generating [SBOM](https://www.ntia.gov/page/software-bill-materials)s for shipped software artifacts). 


## Stability

This is a fundamental module of the Bazel ecosystem that most, if not all, other Bazel modules depend on. Thus stability is very important, and we promise to never change the public API.

> **IMPORTANT**: This module is currently under active development, and not all exported symbols are covered by this stability guarantee. Please refer to the documentation of rules and providers for whether they are currently considered stable.


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

If you are a module author and want to annotate your module, you will need to take the following steps:

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
        purl = purl.bazel(module_name(), module_version()),
        attributes = [
            # ...
        ],
        visibility = ["//visibility:public"],
    )
    ```

    <!-- TODO(yannic): use PURL builder instead of a format string. -->

  - (optional) Add `attributes` to your `package_metadata` target(s).

    `package_metadata` itself only provides information about the identity of a module or package and where it was retrieved from. Additional metadata is provided as `attributes` to the `package_metadata` target (e.g., the license the packages are under, ...).

    > Definition of attributes are currently under development and not ready for wider usage yet. Please avoid adding attributes to OSS modules for now.

  - Annotate all targets with `package_metadata`.

    This step is required for consumers to access the declared metadata.

    There are three options to annotate targets:

      - Add `package_metadata` to all targets individually:

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

      - Add `default_package_metadata` to all packages

        ```starlark
        package(default_package_metadata = ["//:package_metadata"])
        ```

        This provides a simple way to annotate all targets in a package, while preserving the ability to annotate some targets in the package with a different metadata using the method above.

      - Add `default_package_metadata` to `REPO.bazel`

        This method is very similar to adding `default_package_metadata` to all packages, but it requires changing a single file only.

        ```starlark
        repo(default_package_metadata = ["//:package_metadata"])
        ```

        This provides a simple way to annotate all targets in a package, while preserving the ability to annotate some packages or targets in the package with a different metadata using the methods above.

  - Publish your module.

### As an organization

> This is currently under active development.
>
> We will update this page after stabilizing the API.


## API Documentation

### Generic

  - [@package_metadata//:defs.bzl](./defs.md)

#### Providers

  - [@package_metadata//providers:package_attribute_info.bzl](./providers/package_attribute_info.md)
  - [@package_metadata//providers:package_metadata_info.bzl](./providers/package_metadata_info.md)

#### Rules

  - [@package_metadata//rules:package_metadata.bzl](./rules/package_metadata.md)

#### Utils

  - [@package_metadata//purl:purl.bzl](./purl/purl.md)


### Licenses

  - [@package_metadata//licenses:defs.bzl](./licenses/defs.md)

#### Providers

  - [@package_metadata//licenses/providers:license_kind_info.bzl](./licenses/providers/license_kind_info.md)

#### Rules

  - [@package_metadata//licenses/rules:license.bzl](./licenses/rules/license.md)
  - [@package_metadata//licenses/rules:license_kind.bzl](./licenses/rules/license_kind.md)
