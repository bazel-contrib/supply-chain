"""Public API of `@package_metadata`."""

load("//private/providers:package_attribute_info.bzl", _PackageAttributeInfo = "PackageAttributeInfo")
load("//private/providers:package_metadata_info.bzl", _PackageMetadataInfo = "PackageMetadataInfo")
load("//private/rules:package_metadata.bzl", _package_metadata = "package_metadata")

# Providers.
PackageAttributeInfo = _PackageAttributeInfo
PackageMetadataInfo = _PackageMetadataInfo

# Rules
package_metadata = _package_metadata
