"""Declares provider `PackageAttributeOverrideInfo`."""

visibility("public")

def _init(package, overrides = []):
    return {
        "overrides": {o.kind: o for o in overrides},
        "package": package,
    }

PackageAttributeOverrideInfo, _create = provider(
    doc = """
Provider for declaring overrides for attributes of `package_metadata` targets.
""".strip(),
    fields = {
        "overrides": """
""".strip(),
        "package": """
The [Label](https://bazel.build/rules/lib/builtins/Label) of the
`package_metadata` target to override attributes of.
""".strip(),
    },
    init = _init,
)
