"""Public API of `@package_metadata//attributes`."""

load("//attributes:copyright_notice.bzl", _copyright_notice = "copyright_notice")
load("//attributes:cpe.bzl", _cpe = "cpe")

visibility("public")

# Rules
copyright_notice = _copyright_notice
cpe = _cpe
