"""Declares provider `LicenseExceptionInfo`."""

visibility("public")

def _init(identifier, name, *, deprecation = None):
    return {
        "identifier": identifier,
        "name": name,
        "deprecation": deprecation,
    }

LicenseExceptionInfo, _create = provider(
    doc = """
Provides information to identify a license exception.
""".strip(),
    fields = {
        "identifier": """
A [string](https://bazel.build/rules/lib/core/string) uniquely identifying the
license exception (e.g., `Classpath-exception-2.0`).
""".strip(),
        "name": """
A [string](https://bazel.build/rules/lib/core/string) containing the (human
readable) name of the license exception.
""".strip(),
        "deprecation": """
An optional [string](https://bazel.build/rules/lib/core/string) explaining why
the license exception is deprecated.
""".strip(),
    },
    init = _init,
)
