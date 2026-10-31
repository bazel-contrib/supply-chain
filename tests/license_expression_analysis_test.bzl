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
    env.expect.that_bool(mit.kind.exception).equals(False)

    with_node = root.operands[1]
    env.expect.that_str(with_node.operator).equals("WITH")
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
    env.expect.that_str(root.operator).equals("OR")

    mit = root.operands[0]
    env.expect.that_str(mit.kind.identifier).equals("MIT")

    with_node = root.operands[1]
    env.expect.that_str(with_node.operator).equals("WITH")
    env.expect.that_str(with_node.exception.identifier).equals("AdditionRef-acme-linking")
    env.expect.that_bool(with_node.exception.exception).equals(True)
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
