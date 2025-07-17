"""Utils for [purl](https://github.com/package-url/purl-spec)'s `percent encoding`.

Spec: https://github.com/package-url/purl-spec/blob/main/PURL-SPECIFICATION.rst#character-encoding
"""

load("//purl:string.bzl", "string")
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

def percent_encode(value):
    """Encodes the provided string.

    Args:
      value (string): The string to encode.
    Returns:
      The encoded string.
    """

    return "".join([_encode_byte(b) for b in string.to_bytes(value)])
