"""Utils for working with Starlark [string](https://bazel.build/rules/lib/core/string)s."""

load("//purl/private/strings:ascii.bzl", "ascii")
load("//purl/private/strings:bytes.bzl", "bytes")

visibility([
    "//purl/private/...",
])

strings = struct(
    ascii = ascii,
    bytes = bytes,
)
