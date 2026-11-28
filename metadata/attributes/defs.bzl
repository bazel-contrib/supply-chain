"""Public API of `@package_metadata//attributes`."""

load("//attributes:integrity.bzl", _integrity = "integrity")
load("//attributes:copyright_notice.bzl", _copyright_notice = "copyright_notice")
load("//attributes:component_type.bzl", _component_type = "component_type")
load("//attributes:cpe.bzl", _cpe = "cpe")
load("//attributes:patch.bzl", _patch = "patch")

visibility("public")

# Rules
integrity = _integrity
copyright_notice = _copyright_notice
component_type = _component_type
cpe = _cpe
patch = _patch
