"""Declares provider `LicenseExpressionInfo`."""

visibility("public")

LicenseExpressionInfo, _create = provider(
    doc = """
A node in the graph of a parsed
[SPDX license expression](https://spdx.github.io/spdx-spec/v3.0.1/annexes/spdx-license-expressions/).

Leaf nodes reference a `LicenseKindInfo`; inner nodes combine other nodes with
`AND`, `OR` or `WITH`.
""".strip(),
    fields = {
        "operator": """
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
        "or_later": """
For a leaf node, a [bool](https://bazel.build/rules/lib/core/bool) indicating
that the license was followed by `+` in the expression (e.g., `GPL-3.0+`),
meaning that the license or any later version applies. `kind` is the license
as written, not a resolved `-or-later` identifier.
""".strip(),
        "exception": """
For `WITH` nodes, the `LicenseKindInfo` of the license exception (e.g.,
`Classpath-exception-2.0`); otherwise `None`.
""".strip(),
    },
    init = lambda *, operator = None, operands = [], kind = None, or_later = False, exception = None: {
        "operator": operator,
        "operands": operands,
        "kind": kind,
        "or_later": or_later,
        "exception": exception,
    },
)
