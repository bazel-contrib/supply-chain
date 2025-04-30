"""Public API of `@package_metadata`."""

load("//providers:package_attribute_info.bzl", _PackageAttributeInfo = "PackageAttributeInfo")
load("//providers:package_attribute_override_info.bzl", _PackageAttributeOverrideInfo = "PackageAttributeOverrideInfo")
load("//providers:package_metadata_info.bzl", _PackageMetadataInfo = "PackageMetadataInfo")
load("//providers:package_metadata_toolchain_info.bzl", _PackageMetadataToolchainInfo = "PackageMetadataToolchainInfo")
load("//rules:package_metadata.bzl", _package_metadata = "package_metadata")
load("//rules:package_metadata_toolchain.bzl", _package_metadata_toolchain = "package_metadata_toolchain")

visibility("public")

# Providers.
PackageAttributeInfo = _PackageAttributeInfo
PackageAttributeOverrideInfo = _PackageAttributeOverrideInfo
PackageMetadataInfo = _PackageMetadataInfo
PackageMetadataToolchainInfo = _PackageMetadataToolchainInfo

# Rules
package_metadata = _package_metadata
package_metadata_toolchain = _package_metadata_toolchain

# Skipped rules:
#  - package_attribute_override: requires `macro()`, which requires Bazel 8.
