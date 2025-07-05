"""Utils for [purl](https://github.com/package-url/purl-spec)'s `percent encoding`.

Spec: https://github.com/package-url/purl-spec/blob/main/PURL-SPECIFICATION.rst#character-encoding
"""

load("//purl:tables.bzl", "percent_encoding")

visibility([
    "//purl/...",
])

def _encode_byte(b):
    """Encodes a single byte.

    Args:
      c: The byte to encode.
    Returns:
      The encoded string.
    """

    encoded = percent_encoding.encode.get(b, None)
    if not encoded:
        fail("Cannot encode {} (type={})".format(b, type(b)))

    return encoded

def _to_bytes(string):
    """Converts a string to a list of bytes."""

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
    return [hash(c) for c in string.elems()]

def percent_encode(string):
    """Encodes the provided string.

    Args:
      string: The string to encode.
    Returns:
      The encoded string.
    """

    return "".join([_encode_byte(b) for b in _to_bytes(string)])
