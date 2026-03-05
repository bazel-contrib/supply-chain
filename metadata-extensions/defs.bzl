"""Public API of `@package_metadata_extensions`."""

load("//rules:package_metadata_override.bzl", _package_metadata_override = "package_metadata_override")

visibility("public")

# Rules
package_metadata_override = _package_metadata_override
