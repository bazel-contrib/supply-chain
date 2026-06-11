"""Rule for testing `bytes.from_string` and `bytes.to_string`."""

load("//purl/private/strings:bytes.bzl", "bytes")

visibility([
    "//purl/private/strings/...",
])

_bash_executable = """
#!/usr/bin/env bash

echo "All tests passed"
""".strip()

_bat_executable = """
echo "All tests passed"
""".strip()

def _parse_byte_list(byte_string):
    """Parses a comma-separated string of byte values into a list of ints."""
    if not byte_string:
        return []
    return [int(b) for b in byte_string.split(",")]

def _bytes_test_impl(ctx):
    # Test from_string
    for test_string, expected_bytes_str in ctx.attr.from_string_cases.items():
        expected_bytes = _parse_byte_list(expected_bytes_str)
        actual_bytes = bytes.from_string(test_string)
        if expected_bytes != actual_bytes:
            fail("Error in from_string({}): expected {}, got {}".format(
                repr(test_string),
                expected_bytes,
                actual_bytes,
            ))

    # Test to_string
    for expected_string, byte_list_str in ctx.attr.to_string_cases.items():
        byte_list = _parse_byte_list(byte_list_str)
        actual_string = bytes.to_string(byte_list)
        if expected_string != actual_string:
            fail("Error in to_string({}): expected {}, got {}".format(
                byte_list,
                repr(expected_string),
                repr(actual_string),
            ))

    # Test round-trip: from_string -> to_string
    for test_string in ctx.attr.roundtrip_cases:
        byte_list = bytes.from_string(test_string)
        result_string = bytes.to_string(byte_list)
        if test_string != result_string:
            fail("Round-trip failed for {}: got {} (via bytes {})".format(
                repr(test_string),
                repr(result_string),
                byte_list,
            ))

    # Unix does not care about the file extension, so always use `.bat` so it
    # also works on Windows.
    executable = ctx.actions.declare_file("{}.bat".format(ctx.attr.name))
    ctx.actions.write(
        output = executable,
        content = _bash_executable if (ctx.configuration.host_path_separator == ":") else _bat_executable,
        is_executable = True,
    )

    return [
        DefaultInfo(
            files = depset(
                direct = [
                    executable,
                ],
            ),
            executable = executable,
        ),
    ]

bytes_test = rule(
    implementation = _bytes_test_impl,
    attrs = {
        "from_string_cases": attr.string_dict(
            doc = "Dictionary of test_string -> comma-separated byte values",
        ),
        "roundtrip_cases": attr.string_list(
            doc = "List of strings to test for round-trip conversion",
        ),
        "to_string_cases": attr.string_dict(
            doc = "Dictionary of expected_string -> comma-separated byte values",
        ),
    },
    test = True,
)
