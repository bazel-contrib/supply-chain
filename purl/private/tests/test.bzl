"""Test runner for PURL spec tests."""

load("//private:builder.bzl", "build")
load("//private:parser.bzl", "parse")
load("//private/tests:spec.bzl", "tests")
load(
    "//private/tests:spec.custom.bzl",
    "custom_tests",
)

visibility([
    "//private/tests/...",
])

_bash_executable = """
#!/usr/bin/env bash

echo '{message}'
exit {status}
""".strip()

_bat_executable = """
echo '{message}'
exit /b {status}
""".strip()

def _check_build_test(test, failures):
    input = dict(test["input"])
    actual, err = build(**input)
    if test["expected_failure"]:
        if err:
            return

        failures.append({
            "description": test["description"],
            "message": "Expected failure: {}".format(test["expected_failure_reason"]),
        })
    else:
        if err:
            failures.append({
                "description": test["description"],
                "message": "Expected no failure, got {}".format(err),
            })
            return
        expected = test["expected_output"]
        if expected != actual:
            failures.append({
                "description": test["description"],
                "message": "Expected {}, got {}".format(expected, actual),
            })

def _check_parse_test(test, failures):
    actual, err = parse(test["input"])
    if test["expected_failure"]:
        if err:
            return

        failures.append({
            "description": test["description"],
            "message": "Expected failure: {}".format(test["expected_failure_reason"]),
        })
    else:
        if err:
            failures.append({
                "description": test["description"],
                "message": "Expected no failure, got {}".format(err),
            })
            return
        expected = test["expected_output"]
        if expected != actual:
            failures.append({
                "description": test["description"],
                "message": "Expected {}, got {}".format(expected, actual),
            })

def _roundtrip(input):
    components, err = parse(input)
    if err:
        return None, err

    return build(**components)

def _check_roundtrip_test(test, failures):
    actual, err = _roundtrip(test["input"])
    if test["expected_failure"]:
        if err:
            return

        failures.append({
            "description": test["description"],
            "message": "Expected failure: {}".format(test["expected_failure_reason"]),
        })
    else:
        if err:
            failures.append({
                "description": test["description"],
                "message": "Expected no failure, got {}".format(err),
            })
            return
        expected = test["expected_output"]
        if expected != actual:
            failures.append({
                "description": test["description"],
                "message": "Expected {}, got {}".format(expected, actual),
            })

def _purl_spec_test_impl(ctx):
    failures = []
    all_tests = tests + custom_tests
    for test in all_tests:
        if test["test_type"] == "build":
            _check_build_test(test, failures)
        elif test["test_type"] == "parse":
            _check_parse_test(test, failures)
        elif test["test_type"] == "roundtrip":
            _check_roundtrip_test(test, failures)
        else:
            fail("Unexpected test type {}".format(test["test_type"]))

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

purl_spec_test = rule(
    implementation = _purl_spec_test_impl,
    test = True,
)
