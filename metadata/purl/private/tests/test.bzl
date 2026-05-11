"""Test runner for PURL spec tests."""

load("//purl/private:builder.bzl", "build")
load("//purl/private:parser.bzl", "parse")
load("//purl/private/tests:spec.bzl", "tests")
load(
    "//purl/private/tests:spec.custom.bzl",
    "custom_tests",
)

visibility([
    "//purl/private/tests/...",
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

_COMMON_QUALIFIERS = {
    "checksum": True,
    "download_url": True,
    "file_name": True,
    "repository_url": True,
    "vcs_url": True,
    "vers": True,
}

def _split_once_left(value, separator):
    parts = value.split(separator)
    if len(parts) == 1:
        return value, None
    return parts[0], separator.join(parts[1:])

def _split_once_right(value, separator):
    parts = value.split(separator)
    if len(parts) == 1:
        return value, None
    return separator.join(parts[:-1]), parts[-1]

def _type_and_qualifier_keys(input):
    if type(input) == type({}):
        qualifiers = input.get("qualifiers")
        return input.get("type"), qualifiers.keys() if qualifiers else []

    purl, _ = _split_once_right(input, "#")
    purl, qualifiers = _split_once_right(purl, "?")
    if qualifiers == None:
        return None, []

    _, remainder = _split_once_left(purl, ":")
    if remainder == None:
        return None, []
    remainder = remainder.lstrip("/")
    purl_type, _ = _split_once_left(remainder, "/")

    keys = []
    for pair in qualifiers.split("&"):
        key, _ = _split_once_left(pair, "=")
        keys.append(key.lower())
    return purl_type.lower(), keys

def _strict(test):
    if test["test_group"] != "base":
        return False

    _, keys = _type_and_qualifier_keys(test["input"])
    for key in keys:
        if key not in _COMMON_QUALIFIERS:
            return False
    return True

def _check_build_test(test, failures):
    actual, err = build(strict = _strict(test), **test["input"])
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
    actual, err = parse(test["input"], strict = _strict(test))
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

def _check_roundtrip_test(test, failures):
    parsed, err = parse(test["input"], strict = _strict(test))
    if test["expected_failure"]:
        if err:
            return

        _, err = build(strict = _strict(test), **parsed)
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
                "message": "Expected no parse failure, got {}".format(err),
            })
            return
        actual, err = build(strict = _strict(test), **parsed)
        if err:
            failures.append({
                "description": test["description"],
                "message": "Expected no build failure, got {}".format(err),
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
        if test["test_group"] not in ["base", "advanced"]:
            fail("Unexpected test group {}".format(test["test_group"]))
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
