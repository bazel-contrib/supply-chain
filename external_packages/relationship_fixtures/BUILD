load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
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
