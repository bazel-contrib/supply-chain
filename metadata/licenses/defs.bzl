"""Public API of `@package_metadata//licenses`."""

load("//licenses/private/rules:license.bzl", _license = "license")

# Providers.

# Rules
license = _license
