load("//licenses/rules:license_exception.bzl", "license_exception")
load("//licenses/rules:license_kind.bzl", "license_kind")
load(":licenses.bzl", "EXCEPTIONS", "LICENSES")

exports_files(
    ["licenses.bzl"],
    visibility = ["//visibility:public"],
)

# License kinds are named after the lower case identifier. An alias with the
# official identifier is declared when it differs, for backward compatibility.
[
    (
        license_kind(
            name = key,
            deprecation = data.get("deprecation"),
            full_name = data["name"],
            identifier = data["identifier"],
            visibility = ["//visibility:public"],
        ),
        alias(
            name = data["identifier"],
            actual = ":" + key,
            visibility = ["//visibility:public"],
        ) if data["identifier"] != key else None,
    )
    for key, data in LICENSES.items()
]

[
    license_exception(
        name = key,
        deprecation = data.get("deprecation"),
        full_name = data["name"],
        identifier = data["identifier"],
        visibility = ["//visibility:public"],
    )
    for key, data in EXCEPTIONS.items()
]
