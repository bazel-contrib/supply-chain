"""Public API of `@package_metadata_extensions`."""

load("//rules:package_metadata_override.bzl", _package_metadata_override = "package_metadata_override")
load("//providers:package_linking_info.bzl", _PackageLinkingInfo = "PackageLinkingInfo")

visibility("public")

# Rules
package_metadata_override = _package_metadata_override

# Providers
PackageLinkingInfo = _PackageLinkingInfo
