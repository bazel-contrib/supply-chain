"""Utils for working with Starlark [string](https://bazel.build/rules/lib/core/string)s."""

def _to_bytes(value):
    """Converts a string to a list of bytes.

    Args:
      value (string): The string to convert.

    Returns:
      A sequence of bytes in Bazel's native encoding.
    """

    forced_utf8_encoded_string = "".join(value.elems())

    # Bazel reads `BUILD` and `.bzl` files as `ISO-8859-1` (a.k.a., `latin-1`),
    # which is always 1 byte wide. This means that `utf-8` multi-byte characters
    # will end up being multiple characters in a Starlark string (e.g., `ü`,
    # which is unicode `\u00FC`, becomes two invalid `utf-8` bytes
    # `[0xC3, 0xBC]`).
    #
    # In Bazel, Starlark's `hash()` function is implemented using
    # `String#hashCode()`, which hashes with this formula:
    # [s[0]*31^(n-1) + s[1]*31^(n-2) + ... + s[n-1]](https://docs.oracle.com/javase/8/docs/api/java/lang/String.html#hashCode--)
    # (where `s` is the `char[]` internally used by `String`). For one-char
    # `Strings` this means that
    # `String#hashCode() == (int) String#toCharArray()`.
    return [hash(c) for c in forced_utf8_encoded_string.elems()]

string = struct(
    to_bytes = _to_bytes,
)
