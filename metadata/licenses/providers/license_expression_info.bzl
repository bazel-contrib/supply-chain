"""Declares provider `LicenseExpressionInfo`."""

visibility("public")

LicenseExpressionInfo, _create = provider(
    doc = """
A node in the graph of a parsed
[SPDX license expression](https://spdx.github.io/spdx-spec/v2.3.1/SPDX-license-expressions/).

Leaf nodes reference a `LicenseKindInfo`; inner nodes combine other nodes with
`AND`, `OR` or `WITH`.
""".strip(),
    fields = {
        "operand": """
`None` for a leaf node, otherwise the string `AND`, `OR` or `WITH`.
""".strip(),
        "operands": """
A [list](https://bazel.build/rules/lib/core/list) of `LicenseExpressionInfo`:
at least two for `AND`/`OR`, exactly one (the licensed leaf) for `WITH`, empty
for a leaf node.
""".strip(),
        "kind": """
The `LicenseKindInfo` of a leaf node, otherwise `None`.
""".strip(),
        "exception": """
For `WITH` nodes, the `LicenseKindInfo` of the license exception (e.g.,
`Classpath-exception-2.0`); otherwise `None`.
""".strip(),
    },
    init = lambda *, operand = None, operands = [], kind = None, exception = None: {
        "operand": operand,
        "operands": operands,
        "kind": kind,
        "exception": exception,
    },
)
