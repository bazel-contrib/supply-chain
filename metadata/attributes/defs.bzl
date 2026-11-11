"""Public API of `@package_metadata//attributes`."""

load("//attributes:integrity.bzl", _integrity = "integrity")
load("//attributes:copyright_notice.bzl", _copyright_notice = "copyright_notice")
load("//attributes:cpe.bzl", _cpe = "cpe")

visibility("public")

# Rules
integrity = _integrity
copyright_notice = _copyright_notice
cpe = _cpe
