"""Utils for [purl](https://github.com/package-url/purl-spec)'s `percent encoding`.

Spec: https://github.com/package-url/purl-spec/blob/main/PURL-SPECIFICATION.rst#character-encoding
"""

load("//purl/private/percent_encoding:tables.bzl", "encode_byte")
load("//purl/private/strings:strings.bzl", "strings")

visibility([
    "//purl/private/...",
])

def _encode_byte(b):
    """Encodes a single byte.

    Args:
      c: The byte to encode.
    Returns:
      The encoded string.
    """

    encoded = encode_byte.get(b, None)
    if not encoded:
        fail("Cannot encode {} (type={})".format(b, type(b)))

    return encoded

def percent_encode(value):
    """Encodes the provided string.

    Args:
      value (string): The string to encode.
    Returns:
      The encoded string.
    """

    return "".join([_encode_byte(b) for b in strings.to_bytes(value)])
