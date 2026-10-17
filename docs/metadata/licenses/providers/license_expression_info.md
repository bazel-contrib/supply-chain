<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares provider `LicenseExpressionInfo`.

<a id="LicenseExpressionInfo"></a>

## LicenseExpressionInfo

<pre>
load("@package_metadata//licenses/providers:license_expression_info.bzl", "LicenseExpressionInfo")

LicenseExpressionInfo(*, <a href="#LicenseExpressionInfo-operand">operand</a>, <a href="#LicenseExpressionInfo-operands">operands</a>, <a href="#LicenseExpressionInfo-kind">kind</a>, <a href="#LicenseExpressionInfo-or_later">or_later</a>, <a href="#LicenseExpressionInfo-exception">exception</a>)
</pre>

A node in the graph of a parsed
[SPDX license expression](https://spdx.github.io/spdx-spec/v2.3.1/SPDX-license-expressions/).

Leaf nodes reference a `LicenseKindInfo`; inner nodes combine other nodes with
`AND`, `OR` or `WITH`.

**FIELDS**

| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="LicenseExpressionInfo-operand"></a>operand | `None` for a leaf node, otherwise the string `AND`, `OR` or `WITH`. | `None` |
| <a id="LicenseExpressionInfo-operands"></a>operands | A [list](https://bazel.build/rules/lib/core/list) of `LicenseExpressionInfo`: at least two for `AND`/`OR`, exactly one (the licensed leaf) for `WITH`, empty for a leaf node. | `[]` |
| <a id="LicenseExpressionInfo-kind"></a>kind | The `LicenseKindInfo` of a leaf node, otherwise `None`. | `None` |
| <a id="LicenseExpressionInfo-or_later"></a>or_later | Whether a leaf node carries the `+` ("or later") operator. | `False` |
| <a id="LicenseExpressionInfo-exception"></a>exception | For `WITH` nodes, the `LicenseKindInfo` of the license exception (e.g., `Classpath-exception-2.0`); otherwise `None`. | `None` |


