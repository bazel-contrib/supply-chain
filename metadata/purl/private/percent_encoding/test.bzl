"""Rule for testing `percent_encode` and `percent_decode`."""

load("//purl/private/percent_encoding:percent_encoding.bzl", "percent_encode")

visibility([
    "//purl/private/percent_encoding/...",
])

_bash_executable = """
#!/usr/bin/env bash

echo "All tests passed"
""".strip()

_bat_executable = """
echo "Hello World"
""".strip()

def _percent_encoding_test_impl(ctx):
    for decoded, encoded in ctx.attr.cases.items():
        actual_encoded = percent_encode(decoded)
        if encoded != actual_encoded:
            fail("Error encoding {}: expected {}, got {}".format(decoded, encoded, actual_encoded))

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

percent_encoding_test = rule(
    implementation = _percent_encoding_test_impl,
    attrs = {
        "cases": attr.string_dict(
            mandatory = True,
        ),
    },
    test = True,
)
