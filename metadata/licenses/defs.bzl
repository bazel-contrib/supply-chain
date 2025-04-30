"""Public API of `@package_metadata//licenses`."""

load("//licenses/providers:license_kind_info.bzl", _LicenseKindInfo = "LicenseKindInfo")
load("//licenses/rules:license_kind.bzl", _license_kind = "license_kind")

visibility("public")

# Providers.
LicenseKindInfo = _LicenseKindInfo

# Rules
license_kind = _license_kind
