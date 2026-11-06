"""Analysis tests for `license` targets declared with an `expression`."""

load("@package_metadata//licenses:defs.bzl", "LicenseExpressionInfo")
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@rules_testing//lib:analysis_test.bzl", "analysis_test")
load("@rules_testing//lib:truth.bzl", "matching")

def license_expression_failure_test(*, name, message, expression = None, kind = None, kinds = None):
    """Asserts that analyzing a `license` with `expression` or `kind` fails with `message`."""
    license(
        name = name + "_license",
        expression = expression,
        kind = kind,
        kinds = kinds,
        tags = ["manual"],
    )
    analysis_test(
        name = name,
        impl = lambda env, target: _failure_impl(env, target, message),
        target = name + "_license",
        expect_failure = True,
    )

def _failure_impl(env, target, message):
    env.expect.that_target(target).failures().contains_predicate(
        matching.str_matches("*" + message + "*"),
    )

def license_expression_graph_test(*, name, expression, kinds = None):
    """Asserts the shape of the `LicenseExpressionInfo` graph of a fixed expression."""
    license(
        name = name + "_license",
        expression = expression,
        kinds = kinds,
        tags = ["manual"],
    )
    analysis_test(
        name = name,
        impl = _graph_impl,
        target = name + "_license",
    )

def _graph_impl(env, target):
    root = target[LicenseExpressionInfo]
    env.expect.that_str(root.operator).equals("OR")
    env.expect.that_int(len(root.operands)).equals(2)

    mit = root.operands[0]
    env.expect.that_str(mit.kind.identifier).equals("MIT")

    with_node = root.operands[1]
    env.expect.that_str(with_node.operator).equals("WITH")
    env.expect.that_str(with_node.exception.identifier).equals("Classpath-exception-2.0")
    env.expect.that_int(len(with_node.operands)).equals(1)

    gpl = with_node.operands[0]
    env.expect.that_str(gpl.kind.identifier).equals("GPL-2.0-only")

def license_expression_custom_kinds_test(*, name, expression, kinds):
    """Asserts a custom license and custom exception are resolved through `kinds`."""
    license(
        name = name + "_license",
        expression = expression,
        kinds = kinds,
        tags = ["manual"],
    )
    analysis_test(
        name = name,
        impl = _custom_kinds_impl,
        target = name + "_license",
    )

def _custom_kinds_impl(env, target):
    root = target[LicenseExpressionInfo]
    env.expect.that_str(root.operator).equals("OR")

    mit = root.operands[0]
    env.expect.that_str(mit.kind.identifier).equals("MIT")

    with_node = root.operands[1]
    env.expect.that_str(with_node.operator).equals("WITH")
    env.expect.that_str(with_node.exception.identifier).equals("AdditionRef-acme-linking")
    env.expect.that_str(with_node.operands[0].kind.identifier).equals("LicenseRef-acme")

def license_expression_or_later_test(*, name, expression):
    """Asserts `<id>+` is kept as a flag on the license it follows."""
    license(
        name = name + "_license",
        expression = expression,
        tags = ["manual"],
    )
    analysis_test(
        name = name,
        impl = _or_later_impl,
        target = name + "_license",
    )

def _or_later_impl(env, target):
    root = target[LicenseExpressionInfo]
    env.expect.that_int(len(root.operands)).equals(3)

    # `GPL-3.0+` is not resolved to `GPL-3.0-or-later`.
    env.expect.that_str(root.operands[0].kind.identifier).equals("GPL-3.0")
    env.expect.that_bool(root.operands[0].or_later).equals(True)

    # An explicit `-or-later` identifier is untouched and not flagged.
    env.expect.that_str(root.operands[1].kind.identifier).equals("LGPL-3.0-or-later")
    env.expect.that_bool(root.operands[1].or_later).equals(False)

    # `+` is valid even if SPDX has no `-or-later` identifier for the license.
    env.expect.that_str(root.operands[2].kind.identifier).equals("CDDL-1.0")
    env.expect.that_bool(root.operands[2].or_later).equals(True)

def license_expression_case_test(*, name, expression):
    """Asserts standard identifiers are matched case-insensitively."""
    license(
        name = name + "_license",
        expression = expression,
        tags = ["manual"],
    )
    analysis_test(
        name = name,
        impl = _case_impl,
        target = name + "_license",
    )

def _case_impl(env, target):
    root = target[LicenseExpressionInfo]
    env.expect.that_str(root.operator).equals("OR")
    env.expect.that_str(root.operands[0].kind.identifier).equals("MIT")

    with_node = root.operands[1]
    env.expect.that_str(with_node.operator).equals("WITH")
    env.expect.that_str(with_node.operands[0].kind.identifier).equals("GPL-2.0-only")
    env.expect.that_str(with_node.exception.identifier).equals("Classpath-exception-2.0")

def _render(root):
    """Renders a `LicenseExpressionInfo` fully parenthesized, e.g. `(MIT OR (A AND B))`."""
    nodes = [root]
    first_operand = []
    index = 0
    for _ in range(1000):
        if index >= len(nodes):
            break
        first_operand.append(len(nodes))
        nodes.extend(nodes[index].operands)
        index += 1

    text = [None] * len(nodes)
    for i in reversed(range(len(nodes))):
        node = nodes[i]
        if node.operator == None:
            text[i] = node.kind.identifier + ("+" if node.or_later else "")
            continue
        operands = [text[first_operand[i] + j] for j in range(len(node.operands))]
        if node.operator == "WITH":
            text[i] = "({} WITH {})".format(operands[0], node.exception.identifier)
        else:
            text[i] = "(" + (" " + node.operator + " ").join(operands) + ")"
    return text[0]

def license_expression_render_test(*, name, expression, expected, kinds = None):
    """Asserts how an expression is parsed, as `expected`, fully parenthesized."""
    license(
        name = name + "_license",
        expression = expression,
        kinds = kinds,
        tags = ["manual"],
    )
    analysis_test(
        name = name,
        impl = lambda env, target: env.expect.that_str(_render(target[LicenseExpressionInfo])).equals(expected),
        target = name + "_license",
    )

_PARSED = {
    "MIT": "MIT",
    "  MIT  ": "MIT",
    "GPL-2.0+": "GPL-2.0+",
    "((MIT))": "MIT",
    "MIT OR Apache-2.0": "(MIT OR Apache-2.0)",
    "MIT or Apache-2.0 and BSD-3-Clause": "(MIT OR (Apache-2.0 AND BSD-3-Clause))",
    "MIT with Classpath-exception-2.0": "(MIT WITH Classpath-exception-2.0)",
    "MIT OR Apache-2.0 OR BSD-3-Clause": "(MIT OR Apache-2.0 OR BSD-3-Clause)",
    "MIT OR Apache-2.0 AND BSD-3-Clause": "(MIT OR (Apache-2.0 AND BSD-3-Clause))",
    "(MIT OR Apache-2.0) AND BSD-3-Clause": "((MIT OR Apache-2.0) AND BSD-3-Clause)",
    "GPL-2.0-only WITH Classpath-exception-2.0 OR MIT": "((GPL-2.0-only WITH Classpath-exception-2.0) OR MIT)",
}

_PARSED_WITH_KINDS = {
    "LicenseRef-acme WITH AdditionRef-acme-linking": "(LicenseRef-acme WITH AdditionRef-acme-linking)",
    "DocumentRef-d:LicenseRef-acme": "LicenseRef-acme",
    "MIT WITH DocumentRef-d:AdditionRef-acme-linking": "(MIT WITH AdditionRef-acme-linking)",
}

_MALFORMED = {
    "   ": "empty license expression",
    "MIT AND": "incomplete license expression",
    "AND MIT": "unexpected 'AND'",
    "MIT MIT": "missing operator before 'MIT'",
    "MIT AND OR Apache-2.0": "unexpected 'OR'",
    "(MIT": "unbalanced '('",
    "MIT)": "unbalanced ')'",
    "()": "unexpected ')'",
    "MIT (Apache-2.0)": "unexpected '('",
    "MIT Or Apache-2.0": "missing operator before 'Or'",
    "MIT AND or Apache-2.0": "unexpected 'OR'",
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

def license_expression_parsing_tests(*, name, kinds):
    """The grammar of the `expression` attribute of `license`.

    Args:
      name: prefix of the tests and name of their test suite.
      kinds: `license_kind` targets of `LicenseRef-acme` and
        `AdditionRef-acme-linking`.
    """
    tests = []
    for i, (expression, expected) in enumerate(_PARSED.items()):
        test = "{}_parsed_{}".format(name, i)
        license_expression_render_test(name = test, expression = expression, expected = expected)
        tests.append(test)
    for i, (expression, expected) in enumerate(_PARSED_WITH_KINDS.items()):
        test = "{}_parsed_with_kinds_{}".format(name, i)
        license_expression_render_test(name = test, expression = expression, expected = expected, kinds = kinds)
        tests.append(test)
    for i, (expression, message) in enumerate(_MALFORMED.items()):
        test = "{}_malformed_{}".format(name, i)
        license_expression_failure_test(name = test, expression = expression, message = message)
        tests.append(test)
    native.test_suite(name = name, tests = tests)
