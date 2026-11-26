load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load("@package_metadata_extensions//rules:package_metadata_override.bzl", "package_metadata_override")
load(":fixtures.bzl", "image_layer")

package(default_visibility = ["//visibility:public"])

package_metadata(
    name = "image_layer_metadata",
    purl = "pkg:generic/other-provider-image-layer@1.0.0",
)

package_metadata(
    name = "payload_metadata",
    purl = "pkg:generic/server-binary@1.0.0",
)

filegroup(
    name = "payload",
    applicable_licenses = [":payload_metadata"],
)

image_layer(
    name = "provider_image_layer",
    srcs = [":payload"],
    applicable_licenses = [":image_layer_metadata"],
)

# An override declared in an external repository: an unqualified pattern names
# this repository, and an apparent repository name uses this repository's
# mapping.
package_metadata_override(
    name = "override",
    metadata = ":payload_metadata",
    targets = [
        "//a/...",
        "@package_metadata//b/...",
    ],
)
