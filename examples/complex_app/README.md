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

Toolchain involvement is declared with the `toolchain_involvement` module
extension. Each provider module declares the involvement for the toolchain it
provides. The demo patches those declarations into the corresponding
provider's `MODULE.bazel` because the upstream modules are not modified here:

- `rules_go`'s `@io_bazel_rules_go//go:toolchain` is build-only for
  `go_binary`.
- The C++ toolchain type `@bazel_tools//tools/cpp:toolchain_type` is embedded
  for the cgo `go_binary`; the Go binary is linked against the LLVM/glibc
  runtime.
- The Python toolchain type
  `@bazel_tools//tools/python:toolchain_type` is embedded for
  `py_image_layer` and the generated `_py_venv_lib` interpreter closure.
- The Java compiler type `@bazel_tools//tools/jdk:toolchain_type` is
  build-involved for Java binaries. The runtime type
  `@bazel_tools//tools/jdk:runtime_toolchain_type` is resolved from the
  target/exec platform constraints above.
- `rules_nodejs`'s `@rules_nodejs//nodejs:toolchain_type` is a host toolchain
  and is build-involved. The distinct
  `@rules_nodejs//nodejs:runtime_toolchain_type` follows configuration: it is
  runtime for the shipped Node service and build-only for the exec-configured
  frontend bundler. It is not overridden by a blanket `js_binary` entry because
  that rule is used in both configurations.

The demo's root module supplies the `single_version_override` patches that add
the provider tags and each provider's dependency on `supply_chain_tools`.
Although registry MODULE files are used to resolve the graph, Bazel 9 applies
root-supplied `single_version_override` patches to those MODULE files as well.
This was verified with `bazel mod graph`: it succeeds and shows
`supply_chain_tools` as a direct dependency under each patched provider
(`llvm`, `rules_go`, `rules_java`, `rules_nodejs`, and `rules_python`). Bazel
documents that MODULE-file patch changes are effective when the patch is
provided by the root module; patches declared by a dependency would not be.
No `archive_override` is needed. The involvement registry contains only
provider/root module declarations; there are no shipped classification
defaults or optional registry override. The existing `include_build_tools`
setting remains enabled so build-only components appear in these reports.
The root module also declares its own custom `openapi_generate_sources` rule's
JDK compiler toolchain as build-involved.

The external source patches declare package metadata in the generated or
upstream package that supplies each toolchain, rather than routing those
declarations through `third_party/BUILD.bazel`. PURL versions come from the
resolved payload: LLVM's `@llvm_version` repository, rules_go's parsed SDK
version from `go.mod`, Node.js's selected `node_version`, rules_python's
`python_version`, and rules_java's `RUNTIME_VERSION`. Their PURLs identify
`llvm/llvm-project`, `golang/go`, `nodejs/node`, `python/cpython`, and OpenJDK;
the host `local-jdk` has no metadata declaration and is not included.

## Packaging relationship edges

The demo applies `rules_img_relationships.patch` and
`rules_pkg_relationships.patch` to the actual pinned module sources. Each patch
adds a private `_supply_chain_module` marker to the rules it classifies and
declares the relationships in that provider's `MODULE.bazel`. The marker is
derived from `Label("//:x").repo_name` and normalized from Bazel's canonical
repository name to the module name. In rules_pkg 1.1.0, the public `pkg_tar`
macro creates a rule whose kind is `pkg_tar_impl`; its real `srcs` and `deps`
attributes are both classified as contents.

The image PURL is `pkg:oci/toolchain-demo@0.1.0`. Separate package metadata
keeps its Go server and server tar as distinct components; the image contains
both the tar and the server binary. The intermediate image-layer target has no
package metadata, so graph normalization composes its `contains` edge into a
direct image-to-tar edge. CycloneDX represents the image-to-tar and
tar-to-server relationships as BOM compositions with `aggregate: "unknown"`
and sorted `assemblies`; these edges do not appear in `dependencies` or as
relationship properties. SPDX 2.3 emits `CONTAINS` relationships and SPDX 3.0.1
emits `contains` relationships for the same hierarchy. The Go server remains
directly installed at `/app/server`; the tar is also included as
`/app/server.tar`.

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

The image and OCI index and their CycloneDX, SPDX 2.3, and SPDX 3.0.1 JSON-LD
reports were built and validated. The OCI index contains two manifests.
CycloneDX 1.6 has 11 components for one image and 12 for the index; the SPDX
reports have 12 and 13 package records respectively, including the root. The
reports have no duplicate PURLs. Repeated PURLs are merged while distinct
relationship kinds are retained. SPDX 2.3 uses `BUILD_TOOL_OF`, `DEPENDS_ON`,
and `CONTAINS`; SPDX 3.0.1 uses lifecycle-scoped `usesTool`, `dependsOn`, and
`contains` relations.

| Component | Scope | Best involvement |
| --- | --- | --- |
| OCI toolchain-demo image (`pkg:oci/toolchain-demo@0.1.0`) | In scope | Runtime |
| Go server tar | In scope | Runtime |
| Go server | In scope | Runtime |
| Ubuntu 24.04 base image | In scope | Runtime |
| glibc `2.28` from LLVM's configured glibc repository | In scope | Runtime |
| musl `1.2.6` from LLVM's configured musl repository | In scope | Embedded |
| OpenJDK `21` | Excluded | Build |
| OpenJDK `25` | In scope | Runtime |
| LLVM `23.1.2` | In scope | Embedded |
| LLVM runtimes `23.1.2` | Excluded | Build |
| Go `1.27.4` | Excluded | Build |
| Node.js `22.23.4` | In scope | Runtime |
| CPython `3.14.10` | In scope | Embedded |

These are the index CycloneDX classifications. JDK 21 is only used for
compilation/bootstrap; JDK 25 is the Java service's runtime. Node.js has a
runtime path through the shipped service even though the same PURL is also
used by the build-time frontend bundler. CPython is embedded by the Python
image layer. Go `1.27.4` is build-only, including across both libc variants.

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
`aspect_rules_js` 3.6.1, and `rules_pkg` 1.1.0. These releases were confirmed
present in the Bazel Central Registry.

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
GNU/glibc 2.28 platform. The registries contain module declarations only;
there are no shipped classification defaults and no registry override flag is
required. It uses
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
