"""Utils for converting `string`s to arrays of bytes."""

load("//purl/private/strings:chrord.bzl", "chr", "ord")

visibility([
    "//purl/private/strings/...",
])

def _from_string(value):
    """Converts a string to a list of bytes.

    Args:
      value (string): The string to convert.

    Returns:
      A sequence of bytes in Bazel's native encoding.
    """
    return [ord(c) for c in value.elems()]

def _to_string(byte_list):
    """Converts a list of bytes to a string.

    Args:
      byte_list (list[int]): The list of bytes to convert.

    Returns:
      A string in Bazel's native encoding.
    """
    return "".join([chr(b) for b in byte_list])

bytes = struct(
    from_string = _from_string,
    to_string = _to_string,
)
