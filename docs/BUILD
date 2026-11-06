load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    purl = (purl.builder()
        .type("github")
        .namespace("bazel-contrib")
        .name("supply-chain")
        .version("HEAD")
        .subpath("docs")
        .build()),
)
