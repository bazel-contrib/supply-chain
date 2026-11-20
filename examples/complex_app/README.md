# Complex app

This standalone Bazel module builds a multi-language web application and OCI
images for both glibc and musl Linux. Its `MODULE.bazel` locally overrides
`package_metadata`, `supply-chain-go`, and `supply_chain_tools` with the
corresponding directories in this checkout.

The Go/cgo server embeds the generated frontend and serves an OpenAPI
document generated from `openapi.json`. It also exposes Python and Java
helpers. A separate Node.js service and the Python, Java, and Go programs are
packaged as image layers. The Ubuntu-based OCI manifests include Node.js and
Java in the glibc variant; the musl variant includes the Python helper and
statically linked Go server.

## libc variants

The `//:libc` build setting selects `glibc` (the default) or `musl`.
`//libc_variants:container_glibc` and
`//libc_variants:container_musl` transition the same manifest to each setting;
`//:container_index` combines both manifests. Both describe `linux/amd64`.
Since OCI platform descriptors do not identify libc, a container runtime
cannot choose between them based on libc automatically.

```sh
# Build the default glibc container and its reports.
bazel build //:container //:container_cyclonedx //:container_spdx

# Select musl.
bazel build //:container --//:libc=musl

# Build both variants and the index reports.
bazel build //:container_index
bazel build //:container_index_cyclonedx //:container_index_spdx

# Run the integration tests natively on the host.
bazel test --config=itest //:itest
```

SBOM targets are generated from the selected image or image index using the
repository's SBOM rules. CycloneDX and SPDX targets are available for both
the single container and the index.
