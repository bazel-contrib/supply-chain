"""Test that compares the JSON emitted by a `license` target with a golden file.

Both sides are normalized with `jq -S .` so the golden files can be pretty-printed.
"""

load("@jq.bzl//jq:jq.bzl", "jq")

def _license_json_test_impl(ctx):
    actual = ctx.attr.license[DefaultInfo].files.to_list()
    if len(actual) != 1:
        fail("expected exactly one output file from {}".format(ctx.attr.license.label))

    script = ctx.actions.declare_file(ctx.label.name + ".sh")
    ctx.actions.write(
        output = script,
        content = """#!/bin/sh
set -eu
if ! diff -u "{expected}" "{actual}"; then
  echo "license JSON does not match the golden file" >&2
  exit 1
fi
""".format(expected = ctx.file.expected.short_path, actual = actual[0].short_path),
        is_executable = True,
    )
    return [DefaultInfo(
        executable = script,
        runfiles = ctx.runfiles(files = [ctx.file.expected] + actual),
    )]

_license_json_test = rule(
    implementation = _license_json_test_impl,
    attrs = {
        "expected": attr.label(allow_single_file = True, mandatory = True),
        "license": attr.label(mandatory = True),
    },
    test = True,
)

def license_json_test(name, license, expected, **kwargs):
    """Compares the JSON of a `license` target, normalized with jq, to a golden file.

    Args:
      name: test name.
      license: the `license` target.
      expected: golden file, normalized by jq as well.
      **kwargs: forwarded to the test rule.
    """
    jq(
        name = name + "_actual",
        srcs = [license],
        args = ["-S"],
        filter = ".",
        out = name + "_actual.json",
    )
    jq(
        name = name + "_expected",
        srcs = [expected],
        args = ["-S"],
        filter = ".",
        out = name + "_expected.json",
    )
    _license_json_test(
        name = name,
        expected = ":" + name + "_expected",
        license = ":" + name + "_actual",
        **kwargs
    )
