"""Declares provider `LicenseKindInfo`."""

visibility("public")

def _init(identifier, name, *, deprecated = False, exception = False):
    return {
        "identifier": identifier,
        "name": name,
        "deprecated": deprecated,
        "exception": exception,
    }

LicenseKindInfo, _create = provider(
    doc = """
Provides information to identify a license.
""".strip(),
    fields = {
        "identifier": """
A [string](https://bazel.build/rules/lib/core/string) uniquely identifying the
license (e.g., `Apache-2.0`, `EUPL-1.1`).

This is typically the [SPDX identifier](https://spdx.org/licenses/) of the
license, but may also be a non-standard value (e.g., in case of a commercial
license).
""".strip(),
        "name": """
A [string](https://bazel.build/rules/lib/core/string) containing the (human
readable) name of the license (e.g., `Apache License 2.0`, `European Union
Public License 1.1`)
""".strip(),
        "deprecated": """
A [bool](https://bazel.build/rules/lib/core/bool) indicating that the license
identifier is deprecated (e.g., SPDX `GPL-2.0` in favor of `GPL-2.0-only`).
""".strip(),
        "exception": """
A [bool](https://bazel.build/rules/lib/core/bool) indicating that this is a
license exception (e.g., `Classpath-exception-2.0`), which is only valid on the
right-hand side of `WITH`, rather than a license.
""".strip(),
    },
    init = _init,
)
