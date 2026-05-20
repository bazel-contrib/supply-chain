"""Public API of `@package_metadata`."""

load("//common:common.bzl", _package_metadata_common = "package_metadata_common")
load("//providers:package_attribute_info.bzl", _PackageAttributeInfo = "PackageAttributeInfo")
load("//providers:package_metadata_info.bzl", _PackageMetadataInfo = "PackageMetadataInfo")
load("//providers:package_metadata_override_info.bzl", _PackageMetadataOverrideInfo = "PackageMetadataOverrideInfo")
load("//providers:package_metadata_toolchain_info.bzl", _PackageMetadataToolchainInfo = "PackageMetadataToolchainInfo")
load("//providers:target_metadata_info.bzl", _TargetMetadataInfo = "TargetMetadataInfo")
load("//purl:purl.bzl", _purl = "purl")
load("//rules:package_metadata.bzl", _package_metadata = "package_metadata")

visibility("public")

# Providers.
PackageAttributeInfo = _PackageAttributeInfo
PackageMetadataInfo = _PackageMetadataInfo
PackageMetadataOverrideInfo = _PackageMetadataOverrideInfo
PackageMetadataToolchainInfo = _PackageMetadataToolchainInfo
TargetMetadataInfo = _TargetMetadataInfo

# Rules
package_metadata = _package_metadata

# Utils
package_metadata_common = _package_metadata_common
purl = _purl
