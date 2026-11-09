load("//licenses/rules:license_kind.bzl", "license_kind")
load(":licenses.bzl", "LICENSES")

exports_files(
    ["licenses.bzl"],
    visibility = ["//visibility:public"],
)

# One target per license of the SPDX list, generated from `licenses.bzl`.
[
    license_kind(
        name = data["identifier"],
        full_name = data["name"],
        identifier = data["identifier"],
        visibility = ["//visibility:public"],
    )
    for data in LICENSES.values()
]
