"""Public API of `@package_metadata//attributes`."""

load("//attributes:integrity.bzl", _integrity = "integrity")
load("//attributes:copyright_notice.bzl", _copyright_notice = "copyright_notice")
load("//attributes:cpe.bzl", _cpe = "cpe")
load("//attributes:patch.bzl", _patch = "patch")
load("//attributes:patch_set.bzl", _patch_set = "patch_set")

visibility("public")

# Rules
integrity = _integrity
copyright_notice = _copyright_notice
cpe = _cpe
patch = _patch
patch_set = _patch_set
