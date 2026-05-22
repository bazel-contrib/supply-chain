"""Utils for [purl](https://github.com/package-url/purl-spec)'s `percent encoding`.

Spec: https://github.com/package-url/purl-spec/blob/main/PURL-SPECIFICATION.rst#character-encoding
"""

load("//private/percent_encoding:tables.bzl", "encode_byte")
load("//private/strings:strings.bzl", "strings")

visibility([
    "//private/...",
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

    return "".join([_encode_byte(b) for b in strings.bytes.from_string(value)])

def _hex_value(c):
    if c >= "0" and c <= "9":
        return strings.bytes.from_string(c)[0] - strings.bytes.from_string("0")[0]
    if c >= "A" and c <= "F":
        return strings.bytes.from_string(c)[0] - strings.bytes.from_string("A")[0] + 10
    if c >= "a" and c <= "f":
        return strings.bytes.from_string(c)[0] - strings.bytes.from_string("a")[0] + 10
    return None

def percent_decode(value):
    """Decodes percent-encoded bytes in the provided string.

    Args:
      value (string): The string to decode.

    Returns:
      A tuple of (decoded_string, error). On success, error is None.
    """

    bytes = []
    skip_until = -1
    for i in range(len(value)):
        if i < skip_until:
            continue

        c = value[i]
        if c != "%":
            bytes.append(strings.bytes.from_string(c)[0])
            continue

        if i + 2 >= len(value):
            return None, "Incomplete percent-encoded sequence"

        high = _hex_value(value[i + 1])
        low = _hex_value(value[i + 2])
        if high == None or low == None:
            return None, "Invalid percent-encoded sequence"

        bytes.append((high * 16) + low)
        skip_until = i + 3

    return strings.bytes.to_string(bytes), None
