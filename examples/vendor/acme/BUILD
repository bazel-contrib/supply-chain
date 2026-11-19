# A package with a commercial license.

load("@package_metadata//:defs.bzl", "package_metadata")
load("@package_metadata//attributes:patch.bzl", "patch")
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@rules_cc//cc:cc_library.bzl", "cc_library")

package(
    default_package_metadata = [":package_data"],
    default_visibility = ["//visibility:public"],
)

package_metadata(
    name = "package_data",
    attributes = [
        ":license",
        ":fix_coyote",
        ":fix_cve",
    ],
    purl = "pkg:github/acmecorp/coyote@3.1.4?download_url=https://github.com/acmecorp/coyote/archive/refs/tags/3.1.4.tar.gz",
)

# The default license for an entire package is typically named "license".
license(
    name = "license",
    kind = "//my_org/licenses:acme_corp_paid",
    text = "ACME_LICENSE",
)

patch(
    name = "fix_coyote",
    description = "Fixes an off-by-one in the coyote parser.",
    file = "fix_coyote.patch",
    integrity = ["sha256-ctb+vLu9gd8hOWuyikCrcmpVsvR8rBzQ5oTddWri88A="],
    type = "unofficial",
)

patch(
    name = "fix_cve",
    description = "Backport of the upstream fix.",
    integrity = ["sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="],
    resolves = ["CVE-2024-0001"],
    type = "backport",
    url = "https://example.com/coyote/cve-2024-0001.patch",
)

cc_library(
    name = "acme",
    srcs = ["coyote.cc"],
)
