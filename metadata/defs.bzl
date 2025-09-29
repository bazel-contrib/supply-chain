"""Public API of `@package_metadata`."""

load("//providers:package_attribute_info.bzl", _PackageAttributeInfo = "PackageAttributeInfo")
load("//providers:package_metadata_info.bzl", _PackageMetadataInfo = "PackageMetadataInfo")
load("//providers:package_metadata_override_info.bzl", _PackageMetadataOverrideInfo = "PackageMetadataOverrideInfo")
load("//providers:package_metadata_toolchain_info.bzl", _PackageMetadataToolchainInfo = "PackageMetadataToolchainInfo")
load("//purl:purl.bzl", _purl = "purl")
load("//rules:package_metadata.bzl", _package_metadata = "package_metadata")
load("//rules:package_metadata_toolchain.bzl", _package_metadata_toolchain = "package_metadata_toolchain")

visibility("public")

# Providers.
PackageAttributeInfo = _PackageAttributeInfo
PackageMetadataInfo = _PackageMetadataInfo
PackageMetadataOverrideInfo = _PackageMetadataOverrideInfo
PackageMetadataToolchainInfo = _PackageMetadataToolchainInfo

# Rules
package_metadata = _package_metadata
package_metadata_toolchain = _package_metadata_toolchain

# Utils
purl = _purl
