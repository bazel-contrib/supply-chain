"""Module extension for provider-declared toolchain involvement."""

load(":toolchain_involvement.bzl", "normalize_toolchain_type")

_VALID_INVOLVEMENTS = ["runtime", "embedded", "build"]

def merge_toolchain_involvement_declarations(declarations):
    """Merge declarations, preferring the root module and reporting conflicts.

    Each item is a struct with key, rule_kind, toolchain_type, involvement,
    module_name, is_root, and source fields.
    """
    merged = {}
    conflicts = []
    for declaration in declarations:
        key = declaration.key
        current = merged.get(key)
        if current == None:
            merged[key] = declaration
            continue
        if current.involvement == declaration.involvement:
            if declaration.is_root and not current.is_root:
                merged[key] = declaration
            continue
        if declaration.is_root and not current.is_root:
            merged[key] = declaration
            continue
        if current.is_root and not declaration.is_root:
            continue
        conflicts.append(
            "conflicting toolchain involvement declarations for {}: module '{}' ({}) declares '{}' while module '{}' ({}) declares '{}'; the root module wins when it declares this key".format(
                key,
                current.module_name,
                current.source,
                current.involvement,
                declaration.module_name,
                declaration.source,
                declaration.involvement,
            ),
        )
    return struct(
        declarations = [merged[key] for key in sorted(merged.keys())],
        conflicts = conflicts,
    )

def _quote(value):
    return "\"" + value.replace("\\", "\\\\").replace("\"", "\\\"").replace("\n", "\\n").replace("\r", "\\r") + "\""

def _toolchain_involvements_repo_impl(ctx):
    bzl_label = str(ctx.attr._toolchain_involvement_bzl)
    lines = [
        "load({}, \"toolchain_involvement\", \"toolchain_involvements\")".format(_quote(bzl_label)),
        "",
    ]
    entries = []
    for index in range(len(ctx.attr.rule_kinds)):
        name = "declaration_{}".format(index)
        lines.extend([
            "toolchain_involvement(",
            "    name = {},".format(_quote(name)),
            "    rule_kind = {},".format(_quote(ctx.attr.rule_kinds[index])),
            "    toolchain_type = {},".format(_quote(ctx.attr.toolchain_types[index])),
            "    involvement = {},".format(_quote(ctx.attr.involvements[index])),
            ")",
            "",
        ])
        entries.append(":{}".format(name))

    lines.extend([
        "toolchain_involvements(",
        "    name = \"toolchain_involvements\",",
        "    visibility = [\"//visibility:public\"],",
        "    entries = [{}],".format(", ".join([_quote(entry) for entry in entries])),
        ")",
        "",
    ])
    ctx.file("BUILD.bazel", "\n".join(lines))

_toolchain_involvements_repo = repository_rule(
    implementation = _toolchain_involvements_repo_impl,
    attrs = {
        "rule_kinds": attr.string_list(mandatory = True),
        "toolchain_types": attr.string_list(mandatory = True),
        "involvements": attr.string_list(mandatory = True),
        "_toolchain_involvement_bzl": attr.label(
            allow_single_file = True,
            default = Label("//gather_metadata:toolchain_involvement.bzl"),
        ),
    },
)

def _toolchain_involvement_impl(module_ctx):
    declarations = []
    for module in module_ctx.modules:
        for tag in module.tags.declare:
            if tag.involvement not in _VALID_INVOLVEMENTS:
                fail(
                    "toolchain involvement declared by module '{}' must be one of {}; got '{}'".format(
                        module.name,
                        _VALID_INVOLVEMENTS,
                        tag.involvement,
                    ),
                )
            rule_kind = tag.rule_kind
            if not rule_kind:
                fail("toolchain involvement rule_kind must be '*' or a rule kind")
            source = tag.name or str(tag.toolchain_type)
            toolchain_type = str(tag.toolchain_type)
            declarations.append(struct(
                key = "{}|{}".format(rule_kind, normalize_toolchain_type(toolchain_type)),
                rule_kind = rule_kind,
                toolchain_type = toolchain_type,
                involvement = tag.involvement,
                module_name = module.name,
                is_root = module.is_root,
                source = source,
            ))

    result = merge_toolchain_involvement_declarations(declarations)
    if result.conflicts:
        fail(result.conflicts[0])

    _toolchain_involvements_repo(
        name = "supply_chain_toolchain_involvements",
        rule_kinds = [declaration.rule_kind for declaration in result.declarations],
        toolchain_types = [declaration.toolchain_type for declaration in result.declarations],
        involvements = [declaration.involvement for declaration in result.declarations],
    )

toolchain_involvement = module_extension(
    implementation = _toolchain_involvement_impl,
    tag_classes = {
        "declare": tag_class(attrs = {
            "toolchain_type": attr.label(mandatory = True),
            "involvement": attr.string(mandatory = True),
            "rule_kind": attr.string(default = "*"),
            "name": attr.string(),
        }),
    },
)
