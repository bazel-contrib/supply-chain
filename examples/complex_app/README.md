# Toolchain involvement demo

This standalone Bazel module builds glibc and musl Linux container variants,
an OCI image index containing both manifests, and SBOMs for the index. Its
`MODULE.bazel` locally overrides `supply_chain_tools`,
`supply-chain-go`, and `package_metadata` to the corresponding directories in
this checkout.

## What the image does

- The Go/cgo server embeds `web/index.html` and the JavaScript output produced
  by `js_run_binary`. Node.js runs only while Bazel bundles that frontend.
- A separate `js_binary` is packaged with `js_image_layer`. Its Linux Node.js
  runtime and service launcher are included in the container, demonstrating
  the target-runtime case independently of the build-time frontend bundler.
- `/api/python` starts the Python helper packaged by `rules_py`'s
  `py_image_layer`; the Python interpreter is in the image.
- `/api/java` starts a `rules_java` `java_binary` whose launch files are
  included in the Java image layer, including its selected runtime JDK files.
- `/openapi.json` serves a Go constant generated from `openapi.json` by a
  separate Java build tool (`openapi_codegen_tool`).
- The Go server is built with cgo and the hermetic LLVM compiler. The image
  uses Ubuntu 24.04 as its runtime base; the selected LLVM toolchain targets
  the glibc 2.28 ABI.

## Selecting libc and building the index

The `//:libc` string build setting selects `glibc` (default) or `musl`.
`//libc_variants:container_glibc` and `//libc_variants:container_musl` use
transitions to build the same manifest under both settings; the transition also
selects the matching `rules_python` runtime. `//:container_index` combines
those two manifests, so one invocation builds both variants. The musl Go
executable is statically linked: its ELF has no interpreter or `DT_NEEDED`
entries. The glibc executable uses `/lib64/ld-linux-x86-64.so.2` and has
dynamic dependencies on glibc libraries.

The pinned `rules_python` release provides a Linux musl CPython runtime, so the
Python sidecar is included in both variants. The pinned Node.js distribution
and Java runtime are glibc-only; their layers are included only in the glibc
manifest. The musl server returns HTTP 501 for those omitted sidecars.

Both manifests currently describe `linux/amd64`. OCI platform descriptors do
not identify libc, so a runtime selecting from this index cannot distinguish
the glibc and musl entries automatically. The index is useful for demonstrating
multi-variant SBOM collection, but production publishing should add a
selection mechanism (for example separate tags or variant annotations).

## Toolchain involvement edges

The demo's root `MODULE.bazel` registers explicit Java toolchains, which
take precedence over the `rules_java` registrations:

- The Java compiler toolchain uses the demo's `java_toolchain`, configured for
  source/target 21 and JDK 21.
- The runtime toolchain for the Linux/x86_64 target platform uses JDK 25, which
  is packaged with the Java service and can run Java 21 bytecode.
- The runtime toolchain for the macOS/arm64 exec platform uses JDK 21 for
  exec-configured Java tools.
- The bootstrap runtime toolchain for the macOS/arm64 exec platform also uses
  JDK 21, avoiding rules_java's `local-jdk` fallback.

The compiler uses a direct `toolchain()` wrapper around the `java_toolchain`
target, without the generated `target_settings` that would require a
`--java_language_version` flag. The runtime toolchains are selected by platform
constraints. This removes the language/runtime-version flags from `.bazelrc`.
The demo applies a small root `rules_java` patch to expose its otherwise
private `bootclasspath` helper, and configures both that helper's JDK alias
and its language-version bootstrap runtime to the hermetic JDK 21. Java
targets live in a separate package without the demo root's default package
metadata, so build-helper targets do not collapse into the image component.
The explicit `bootstrap_runtime_toolchain_type` registration prevents
`local-jdk` from being selected.

Toolchain involvement belongs on each rule as the private
`_supply_chain_toolchain_involvement` label-keyed string-dictionary attribute.
For example, the custom `openapi_generate_sources` rule declares its JDK
compiler toolchain as
`Label("@bazel_tools//tools/jdk:toolchain_type"): "build"` in the attribute
default. These keys are real label dependencies, so typos or unavailable labels
fail at loading time; the aspect excludes the private attribute from SBOM
dependency traversal. Third-party rule declarations are patched into their
underlying `rule()` definitions, so the upstream rulesets do not depend on
`supply_chain_tools` or `package_metadata`. The root module has no shared
toolchain-involvement registry or override mechanism; changing a declaration
requires patching its rule. Unannotated rules continue to use the aspect's
configuration-based defaults.

The helper's attrs dictionary is merged with `|`, which requires Bazel 7.2 or
newer. That minimum is compatible with the repository's Bazel 7.x, 8.x, and
9.x CI lines. Because declarations live on the rule, a consuming root module
cannot override them without patching that rule.

The external source patches declare package metadata in the generated or
upstream package that supplies each toolchain, rather than routing those
declarations through `third_party/BUILD.bazel`. PURL versions come from the
resolved payload: LLVM's `@llvm_version` repository, rules_go's parsed SDK
version from `go.mod`, Node.js's selected `node_version`, rules_python's
`python_version`, and rules_java's `RUNTIME_VERSION`. Their PURLs identify
`llvm/llvm-project`, `golang/go`, `nodejs/node`, `python/cpython`, and OpenJDK;
the host `local-jdk` has no metadata declaration and is not included.

## Integration test with rules_itest

`//:itest` uses [rules_itest](https://github.com/hermeticbuild/rules_itest) to start
the Go `server` (which also runs the Python and Java binaries behind `/api/python`
and `/api/java`) and the Node.js service on auto-assigned ports, wait for their
health checks, and run `itest_test.go` against them. It runs on the host platform,
since the container targets linux_x86_64:

```sh
bazel test --config=itest //:itest
bazel build --config=itest //:itest_cyclonedx //:itest_spdx
```

`itest_sbom` takes the test as its target, so the SBOM ties together the Go, Node.js,
Python and Java components with their toolchains and rules_itest itself.

## Verified SBOM components

The image index and its CycloneDX, SPDX 2.3, and SPDX 3.0.1 JSON-LD reports
were built. The OCI index contains two manifests. CycloneDX has ten
third-party components; each SPDX report has eleven package records including
the root, with no duplicate PURLs. Repeated PURLs are merged while distinct
relationship kinds are retained. SPDX 2.3 uses `BUILD_TOOL_OF`, `DEPENDS_ON`,
and `CONTAINS`; SPDX 3.0.1 uses lifecycle-scoped `usesTool`, `dependsOn`, and
`contains` relations.

| Component | Scope | Best involvement |
| --- | --- | --- |
| Ubuntu 24.04 base image | In scope | Runtime |
| glibc `2.28` from LLVM's configured glibc repository | In scope | Runtime |
| musl `1.2.6` from LLVM's configured musl repository | In scope | Embedded |
| OpenJDK `21` | Excluded | Build |
| OpenJDK `25` | In scope | Runtime |
| LLVM `23.1.2` | In scope | Embedded |
| LLVM runtimes `23.1.2` | Excluded | Build |
| Go `1.27.4` | In scope | Embedded |
| Node.js `22.23.4` | In scope | Runtime |
| CPython `3.14.10` | In scope | Embedded |

These are the index CycloneDX classifications. JDK 21 is only used for
compilation/bootstrap; JDK 25 is the Java service's runtime. Node.js has a
runtime path through the shipped service even though the same PURL is also
used by the build-time frontend bundler. CPython is embedded by the Python
image layer. Go `1.27.4` is embedded, including across both libc variants: the
Go runtime and standard library are linked into the server binary. The Go SDK is
one component, so the compiler shares that classification.

The musl executable's static linkage was verified from its ELF (`PT_INTERP`
absent, no `DT_NEEDED` entries), so musl is embedded. The glibc executable
uses `/lib64/ld-linux-x86-64.so.2` and has dynamic glibc dependencies, so
glibc is runtime. Its PURL comes from LLVM's generated `glibc_2.28` repository;
a metadata-only `runtime_dependency` target in that repository connects the
toolchain-selected glibc component to the Go binary without copying toolchain
headers or linker inputs into the image.

## Module versions

The demo pins the requested BCR releases: `rules_img` 0.3.26,
`rules_java` 9.9.0, `rules_py` 2.0.0-alpha.11, `rules_python` 2.5.1,
`rules_go` 0.65.0, `llvm` 0.8.27, `rules_nodejs` 6.7.7, and
`aspect_rules_js` 3.6.1. These releases were confirmed present in the Bazel
Central Registry.

## Toolchain versions and glibc

For glibc, the hermetic `llvm` toolchain provides headers and linker stubs for
`@llvm//platforms:linux_x86_64_gnu.2.28`; the image's Ubuntu base supplies the
dynamic runtime. The glibc PURL comes from LLVM's generated
`@llvm++glibc+glibc_2.28` repository, not an Ubuntu package declaration.
The musl PURL likewise comes from LLVM's generated musl source repository, with
its version derived from the toolchain's musl extension.

The LLVM toolchain requires Go 1.27 or newer for cgo without patches. The
`go_sdk.from_file` extension reads Go 1.27.4 from `go.mod`, so the demo uses
the upstream Go SDK directly without the older compiler patch. Toolchain
archives and JDK, Go, Node, and Python runtimes are multi-gigabyte downloads.
Check free disk before building. The container index, both libc variants, and
all index report targets have been built successfully.

```sh
# The selected standalone image/SBOM (glibc by default):
bazel build //:container //:container_cyclonedx //:container_spdx //:container_spdx_jsonld

# Select the musl variant:
bazel build //:container --//:libc=musl
bazel build //:container_cyclonedx //:container_spdx //:container_spdx_jsonld --//:libc=musl

# Build both variants and SBOMs over their OCI index:
bazel build //:container_index
bazel build //:container_index_sbom //:container_index_cyclonedx //:container_index_spdx //:container_index_spdx_jsonld

# Existing glibc reports:
bazel build //:container_cyclonedx //:container_spdx //:container_spdx_jsonld
```

The `.bazelrc` opts in to build-tool edges and selects the hermetic LLVM
GNU/glibc 2.28 platform. The default registry combines module-extension tags
with shipped defaults; no registry override flag is required. It uses
`--experimental_platform_in_output_dir` for stable output paths. To resolve the
module graph without building the demo, use:

```sh
bazel mod graph
```

Generated reports are written under `bazel-bin/`:

- `container_cyclonedx.json`
- `container_spdx.json` (SPDX 2.3)
- `container_spdx_jsonld.json` (SPDX 3.0.1 JSON-LD)
- `container_index_index.json` (two OCI manifest descriptors)
- `container_index_cyclonedx.json`
- `container_index_spdx.json`
- `container_index_spdx_jsonld.json`
