"""Tests for SPDX license expression parsing."""

load("@package_metadata//licenses/private:expression.bzl", "parse_license_expression", "try_parse_license_expression")

def _lic(id, or_later = False, ref = False):
    return {"id": id, "or_later": or_later, "ref": ref}

def _test_impl(ctx):
    mit = _lic("MIT")
    apache = _lic("Apache-2.0")
    bsd = _lic("BSD-3-Clause")
    cases = {
        "MIT": mit,
        "  MIT  ": mit,
        "GPL-2.0+": _lic("GPL-2.0", or_later = True),
        "LicenseRef-foo.1": _lic("LicenseRef-foo.1", ref = True),
        "DocumentRef-d:LicenseRef-x": _lic("DocumentRef-d:LicenseRef-x", ref = True),
        "MIT OR Apache-2.0": {"operand": "OR", "operands": [mit, apache]},
        "MIT OR Apache-2.0 OR BSD-3-Clause": {"operand": "OR", "operands": [mit, apache, bsd]},
        "MIT OR Apache-2.0 AND BSD-3-Clause": {
            "operand": "OR",
            "operands": [mit, {"operand": "AND", "operands": [apache, bsd]}],
        },
        "(MIT OR Apache-2.0) AND BSD-3-Clause": {
            "operand": "AND",
            "operands": [{"operand": "OR", "operands": [mit, apache]}, bsd],
        },
        "GPL-2.0-only WITH Classpath-exception-2.0 OR MIT": {
            "operand": "OR",
            "operands": [
                {"operand": "WITH", "license": _lic("GPL-2.0-only"), "exception": "Classpath-exception-2.0"},
                mit,
            ],
        },
        "((MIT))": mit,
        "LicenseRef-a WITH AdditionRef-b": {
            "operand": "WITH",
            "license": _lic("LicenseRef-a", ref = True),
            "exception": "AdditionRef-b",
        },
    }
    for expression, expected in cases.items():
        actual = parse_license_expression(expression)
        if actual != expected:
            fail("'{}': expected {}, got {}".format(expression, expected, actual))

    errors = {
        "": "empty license expression",
        "   ": "empty license expression",
        "MIT AND": "incomplete license expression",
        "AND MIT": "unexpected 'AND'",
        "MIT MIT": "missing operator before 'MIT'",
        "MIT AND OR Apache-2.0": "unexpected 'OR'",
        "(MIT": "unbalanced '('",
        "MIT)": "unbalanced ')'",
        "()": "unexpected ')'",
        "MIT (Apache-2.0)": "unexpected '('",
        "mit or apache-2.0": "missing operator before 'or'",
        "MIT_1": "invalid character '_'",
        "MIT, Apache-2.0": "invalid character ','",
        "LicenseRef-x+": "invalid license reference",
        "LicenseRef-": "invalid license reference",
        "DocumentRef-d:MIT": "invalid license reference",
        "(MIT OR Apache-2.0) WITH Classpath-exception-2.0": "'WITH' requires",
        "MIT WITH (Classpath-exception-2.0)": "'WITH' requires",
        "MIT WITH Classpath-exception-2.0+": "'WITH' requires",
        "MIT WITH A WITH B": "'WITH' requires",
        "WITH": "unexpected 'WITH'",
        "MIT WITH LicenseRef-x": "'WITH' requires",
    }
    for expression, message in errors.items():
        actual = try_parse_license_expression(expression)
        if "error" not in actual or message not in actual["error"]:
            fail("'{}': expected error containing '{}', got {}".format(expression, message, actual))

    script = ctx.actions.declare_file(ctx.label.name + ".sh")
    ctx.actions.write(script, "#!/usr/bin/env bash\nexit 0\n", is_executable = True)
    return [DefaultInfo(executable = script)]

license_expression_test = rule(
    implementation = _test_impl,
    test = True,
)
