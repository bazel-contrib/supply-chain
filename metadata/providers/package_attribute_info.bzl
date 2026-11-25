"""Declares provider `PackageAttributeInfo`."""

visibility("public")

def _init(kind, attributes, files = [], allow_multiple = False):
    return {
        "allow_multiple": allow_multiple,
        "attributes": attributes,
        "files": depset(
            direct = [
                attributes,
            ],
            transitive = files,
        ),
        "kind": kind,
    }

PackageAttributeInfo, _create = provider(
    doc = """
Provider for declaring a single attribute of a Bazel package (e.g., the license
it is available under or the copyright notice).

Attributes are attached to a package by passing targets providing this to the
`attributes` of a `package_metadata` target. Declaring one is the extension
point for organizations that need to inject metadata beyond the attributes
provided by `@package_metadata`.
""".strip(),
    fields = {
        "allow_multiple": """
Whether a `package_metadata` target may declare more than one attribute of this
`kind`. Defaults to `False`.

If `False`, declaring two attributes of this `kind` on the same
`package_metadata` fails the build. If `True`, all of them are recorded, in the
order they were declared, and consumers are expected to handle a list of
`attributes` files for this `kind` (e.g., a `patch` attribute declared once per
patch).
""".strip(),
        "attributes": """
The [File](https://bazel.build/rules/lib/builtins/File) containing the
attributes.

The format of this file depends on the `kind` of attribute. Please consult the
documentation of the attribute (e.g., `license` documents the JSON object it
writes).
""".strip(),
        "files": """
A [depset](https://bazel.build/rules/lib/builtins/depset) of
[File](https://bazel.build/rules/lib/builtins/File)s containing information
about this attribute.
""".strip(),
        "kind": """
The identifier of the attribute.

This should generally be in reverse DNS format (e.g., `com.example.foo`).
""".strip(),
    },
    init = _init,
)
