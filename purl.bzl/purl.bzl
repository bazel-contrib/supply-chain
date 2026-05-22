"""Module defining urils for [purl](https://github.com/package-url/purl-spec)s."""

load("//private:builder.bzl", _builder = "builder")
load("//private:parser.bzl", _parse = "parse")

visibility("public")

builder = _builder
parse = _parse
