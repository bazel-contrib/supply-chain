load("@package_metadata//rules:package_attribute_override.bzl", "package_attribute_override")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@package_metadata//rules:package_metadata_toolchain.bzl", "package_metadata_toolchain")

package_metadata(
    name = "package_metadata",
    purl = "pkg:github/bazel-contrib/supply-chain@HEAD#docs",
)

package_attribute_override(
    name = "foo",
    package = "@package_metadata",
)

package_metadata_toolchain(
    name = "bar",
)

toolchain(
    name = "baz",
    toolchain = ":bar",
    toolchain_type = "@package_metadata//toolchains:type",
)
