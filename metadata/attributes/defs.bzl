"""Public API of `@package_metadata//attributes`."""

load("//attributes:copyright_notice.bzl", _copyright_notice = "copyright_notice")
load("//attributes:cpe.bzl", _cpe = "cpe")
load("//attributes:integrity.bzl", _integrity = "integrity")
load("//attributes:originator.bzl", _originator = "originator")
load("//attributes:patch.bzl", _patch = "patch")
load("//attributes:supplier.bzl", _supplier = "supplier")

visibility("public")

# Rules
integrity = _integrity
copyright_notice = _copyright_notice
cpe = _cpe
patch = _patch
supplier = _supplier
originator = _originator
