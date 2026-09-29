"""Tests for public PURL helpers."""

load("@package_metadata//purl:purl.bzl", "purl")

_bash_executable = """
#!/usr/bin/env bash

echo '{message}'
exit {status}
""".strip()

_bat_executable = """
echo '{message}'
exit /b {status}
""".strip()

def _expect_equals(description, actual, expected, failures):
    if actual != expected:
        failures.append({
            "actual": actual,
            "description": description,
            "expected": expected,
        })

def _purl_bazel_registry_test_impl(ctx):
    failures = []

    _expect_equals(
        "default registry is omitted",
        purl.bazel("rules_java", "7.8.0"),
        "pkg:bazel/rules_java@7.8.0",
        failures,
    )
    _expect_equals(
        "custom registry is emitted as repository_url",
        purl.bazel(
            "rules_java",
            "7.8.0",
            registry = "https://example.org/bazel-registry",
        ),
        "pkg:bazel/rules_java@7.8.0?repository_url=https:%2F%2Fexample.org%2Fbazel-registry",
        failures,
    )

    content = _bash_executable if (ctx.configuration.host_path_separator == ":") else _bat_executable

    # Unix does not care about the file extension, so always use `.bat` so it
    # also works on Windows.
    executable = ctx.actions.declare_file("{}.bat".format(ctx.attr.name))
    ctx.actions.write(
        output = executable,
        content = content.format(
            message = json.encode_indent(failures),
            status = 1 if len(failures) else 0,
        ),
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

purl_bazel_registry_test = rule(
    implementation = _purl_bazel_registry_test_impl,
    test = True,
)
