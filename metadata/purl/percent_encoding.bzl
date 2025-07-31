"""Utils for [purl](https://github.com/package-url/purl-spec)'s `percent encoding`.

Spec: https://github.com/package-url/purl-spec/blob/main/PURL-SPECIFICATION.rst#character-encoding
"""

load("//purl:string.bzl", "string")
load("//purl/private:tables.bzl", "encode_byte")

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

    return "".join([_encode_byte(b) for b in string.to_bytes(value)])
