"""Utils for working with Starlark [string](https://bazel.build/rules/lib/core/string)s."""

load("//private/strings:ascii.bzl", "ascii")
load("//private/strings:bytes.bzl", "bytes")

visibility([
    "//private/...",
])

strings = struct(
    ascii = ascii,
    bytes = bytes,
)
