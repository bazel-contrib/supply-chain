"""Analysis tests for `license` targets declared with an `expression`."""

load("@package_metadata//licenses:defs.bzl", "LicenseExpressionInfo")
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@rules_testing//lib:analysis_test.bzl", "analysis_test")
load("@rules_testing//lib:truth.bzl", "matching")

def license_expression_failure_test(*, name, expression, message, kinds = None):
    """Asserts that analyzing a `license` with `expression` fails with `message`."""
    license(
        name = name + "_license",
        expression = expression,
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
    env.expect.that_str(root.operand).equals("OR")
    env.expect.that_int(len(root.operands)).equals(2)

    mit = root.operands[0]
    env.expect.that_str(mit.kind.identifier).equals("MIT")
    env.expect.that_bool(mit.kind.exception).equals(False)

    with_node = root.operands[1]
    env.expect.that_str(with_node.operand).equals("WITH")
    env.expect.that_str(with_node.exception.identifier).equals("Classpath-exception-2.0")
    env.expect.that_bool(with_node.exception.exception).equals(True)
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
    env.expect.that_str(root.operand).equals("OR")

    mit = root.operands[0]
    env.expect.that_str(mit.kind.identifier).equals("MIT")

    with_node = root.operands[1]
    env.expect.that_str(with_node.operand).equals("WITH")
    env.expect.that_str(with_node.exception.identifier).equals("AdditionRef-acme-linking")
    env.expect.that_bool(with_node.exception.exception).equals(True)
    env.expect.that_str(with_node.operands[0].kind.identifier).equals("LicenseRef-acme")

def license_expression_or_later_test(*, name, expression):
    """Asserts `<id>+` resolves to its standalone SPDX identifier."""
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
    env.expect.that_int(len(root.operands)).equals(2)

    # `GPL-3.0+` is the standalone identifier `GPL-3.0-or-later`.
    env.expect.that_str(root.operands[0].kind.identifier).equals("GPL-3.0-or-later")

    # An explicit `-or-later` identifier is untouched.
    env.expect.that_str(root.operands[1].kind.identifier).equals("LGPL-3.0-or-later")
