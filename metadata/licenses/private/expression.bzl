"""Parser for SPDX license expressions.

Implements the grammar of
https://spdx.github.io/spdx-spec/v3.0.1/annexes/spdx-license-expressions/
(SPDX 3.0.1, Annex).

```
idstring = 1*(ALPHA / DIGIT / "-" / "." )
license-id = <short form license identifier from SPDX License List>
license-exception-id = <short form license exception identifier from SPDX License List>
license-ref = [%s"DocumentRef-"(idstring)":"]%s"LicenseRef-"(idstring)
addition-ref = [%s"DocumentRef-"(idstring)":"]%s"AdditionRef-"(idstring)
simple-expression = license-id / license-id"+" / license-ref
addition-expression = license-exception-id / addition-ref
compound-expression = (simple-expression /
  simple-expression ( %s"WITH" / %s"with" ) addition-expression /
  compound-expression ( %s"AND" / %s"and" ) compound-expression /
  compound-expression ( %s"OR" / %s"or" ) compound-expression /
  "(" compound-expression ")" )
license-expression = (simple-expression / compound-expression)
```

As in SPDX, `+` binds tighter than `WITH`, which binds tighter than `AND`,
which binds tighter than `OR`. There must be no whitespace between a
license id and `+`, and whitespace around `WITH`. `AND` and `OR` need
whitespace and/or parentheses around them. In addition, this parser:

* matches `license-id` and `license-exception-id` case-insensitively, as
  SPDX recommends, and the part of a reference after `LicenseRef-` or
  `AdditionRef-` as well;
* does not yet accept the lowercase operators `and`, `or` and `with`, nor a
  `DocumentRef-` prefix on an `addition-ref`;
* requires every `license-ref` and `addition-ref` to be declared as a
  `license_kind` and passed in `kinds`.
"""

visibility("public")

_ID_CHARS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-"
_OPERATORS = {
    "OR": 1,
    "AND": 2,
    "WITH": 3,
}

def _error(message):
    return {"error": message}

def _tokenize(expression):
    tokens = []
    current = ""
    for c in expression.elems():
        if c in " \t\r\n":
            if current:
                tokens.append(current)
                current = ""
        elif c in "()":
            if current:
                tokens.append(current)
                current = ""
            tokens.append(c)
        elif c in _ID_CHARS or c in "+:":
            current += c
        else:
            return _error("invalid character '{}' in license expression '{}'".format(c, expression))
    if current:
        tokens.append(current)
    return tokens

def _is_idstring(s):
    if not s:
        return False
    for c in s.elems():
        if c not in _ID_CHARS:
            return False
    return True

def _parse_operand(token, expression):
    """Parses a `simple-expression` or a `license-exception-id`."""
    or_later = token.endswith("+")
    body = token[:-1] if or_later else token

    if ":" in body:
        # document-ref ":" license-ref
        parts = body.split(":")
        if (len(parts) != 2 or
            not parts[0].startswith("DocumentRef-") or
            not parts[1].startswith("LicenseRef-") or
            not _is_idstring(parts[0][len("DocumentRef-"):]) or
            not _is_idstring(parts[1][len("LicenseRef-"):])):
            return _error("invalid license reference '{}' in license expression '{}'".format(token, expression))
        if or_later:
            return _error("'+' is not allowed after a license reference in '{}'".format(expression))
        return {"id": body, "or_later": False, "ref": True}

    if not _is_idstring(body):
        return _error("invalid license identifier '{}' in license expression '{}'".format(token, expression))

    if body.startswith("LicenseRef-"):
        if or_later or not _is_idstring(body[len("LicenseRef-"):]):
            return _error("invalid license reference '{}' in license expression '{}'".format(token, expression))
        return {"id": body, "or_later": False, "ref": True}

    return {"id": body, "or_later": or_later, "ref": False}

def _apply(operands, op, expression):
    if len(operands) < 2:
        return _error("missing operand for '{}' in license expression '{}'".format(op, expression))
    right = operands.pop()
    left = operands.pop()

    if op == "WITH":
        if "operator" in left or "operator" in right or right["or_later"] or right["ref"]:
            return _error("'WITH' requires a license on the left and an exception id on the right in '{}'".format(expression))
        operands.append({
            "operator": "WITH",
            "license": left,
            "exception": right["id"],
        })
        return

    children = []
    for side in (left, right):
        if side.get("operator") == op:
            children.extend(side["operands"])
        else:
            children.append(side)
    operands.append({"operator": op, "operands": children})

def try_parse_license_expression(expression):
    """Parses an SPDX license expression into a tree.

    Nodes are dicts, one of:

    * A license: `{"id": str, "or_later": bool, "ref": bool}` (no `operator`), where
      `or_later` flags a trailing `+` (which is not part of `id`) and `ref` is
      true for `LicenseRef-`/`DocumentRef-` identifiers.
    * `{"operator": "WITH", "license": <license node>, "exception": str}`
    * `{"operator": "AND" | "OR", "operands": [<node>, ...]}` (at least 2 operands)

    `WITH` binds tighter than `AND`, which binds tighter than `OR`. Operators
    are case-sensitive.

    Args:
      expression: The license expression string.

    Returns:
      The root node of the parsed expression, or `{"error": <message>}` if the
      expression is malformed.
    """
    tokens = _tokenize(expression)
    if type(tokens) == "dict":
        return tokens
    if not tokens:
        return _error("empty license expression")

    operands = []
    ops = []
    expect_operand = True

    for token in tokens:
        if token == "(":
            if not expect_operand:
                return _error("unexpected '(' in license expression '{}'".format(expression))
            if ops and ops[-1] == "WITH":
                return _error("'WITH' requires a license on the left and an exception id on the right in '{}'".format(expression))
            ops.append(token)
        elif token == ")":
            if expect_operand:
                return _error("unexpected ')' in license expression '{}'".format(expression))
            for _ in range(len(ops) + 1):
                if not ops:
                    return _error("unbalanced ')' in license expression '{}'".format(expression))
                top = ops.pop()
                if top == "(":
                    break
                error = _apply(operands, top, expression)
                if error:
                    return error
        elif token in _OPERATORS:
            if expect_operand:
                return _error("unexpected '{}' in license expression '{}'".format(token, expression))
            for _ in range(len(ops)):
                if ops[-1] == "(" or _OPERATORS[ops[-1]] < _OPERATORS[token]:
                    break
                error = _apply(operands, ops.pop(), expression)
                if error:
                    return error
            ops.append(token)
            expect_operand = True
        else:
            if not expect_operand:
                return _error("missing operator before '{}' in license expression '{}'".format(token, expression))
            operand = _parse_operand(token, expression)
            if "error" in operand:
                return operand
            operands.append(operand)
            expect_operand = False

    if expect_operand:
        return _error("incomplete license expression '{}'".format(expression))

    for _ in range(len(ops)):
        op = ops.pop()
        if op == "(":
            return _error("unbalanced '(' in license expression '{}'".format(expression))
        error = _apply(operands, op, expression)
        if error:
            return error

    return operands[0]

def parse_license_expression(expression):
    """Like `try_parse_license_expression`, but fails on malformed expressions."""
    result = try_parse_license_expression(expression)
    if "error" in result:
        fail(result["error"])
    return result

def canonical_id(id, spdx_ids):
    """Returns the canonical spelling of a standard SPDX identifier.

    SPDX matches license and exception identifiers case-insensitively. Other
    identifiers (e.g., `LicenseRef-`) are returned unchanged.

    Args:
      id: A license or exception identifier.
      spdx_ids: Maps the lower case form of each SPDX identifier to its
        canonical spelling.
    """
    return spdx_ids.get(id.lower(), id)

def license_expression_ids(tree, spdx_ids = {}):
    """Returns the sorted, de-duplicated license and exception ids of a parsed expression.

    Args:
      tree: A parsed expression.
      spdx_ids: Maps the lower case form of each SPDX identifier to its
        canonical spelling. Standard identifiers are reported in their
        canonical spelling.
    """
    ids = {}
    stack = [tree]
    for _ in range(100000):
        if not stack:
            break
        node = stack.pop()
        if "operands" in node:
            stack.extend(node["operands"])
        elif node.get("operator") == "WITH":
            stack.append(node["license"])
            ids[canonical_id(node["exception"], spdx_ids)] = True
        else:
            ids[canonical_id(node["id"], spdx_ids)] = True
    return sorted(ids)
