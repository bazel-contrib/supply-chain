"""Module defining urils for [purl](https://github.com/package-url/purl-spec)s."""

load("@purl.bzl", _bazel = "bazel", _builder = "builder", _parse = "parse")

visibility("public")

purl = struct(
    builder = _builder,
    bazel = _bazel,
    parse = _parse,
)
