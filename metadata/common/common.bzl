"""Common utilities for creating package metadata.

These helpers are building blocks for rule authors who need to declare
`PackageMetadataInfo` from their own rules, rather than via the
`package_metadata` rule.
"""

load("//common/private:create_package_metadata.bzl", "create_package_metadata")

visibility("public")

package_metadata_common = struct(
    create_package_metadata = create_package_metadata,
)
