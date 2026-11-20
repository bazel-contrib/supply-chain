"""Module extension for provider-declared toolchain involvement."""

load(":relationships.bzl", "normalize_rule_module")
load(":toolchain_involvement.bzl", "normalize_toolchain_type")

_VALID_INVOLVEMENTS = ["runtime", "embedded", "build"]
_VALID_RELATIONSHIPS = [
    "depends_on",
    "contains",
    "static_link",
    "dynamic_link",
    "test_dependency",
    "dev_dependency",
]

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

def _relationships_repo_impl(ctx):
    bzl_label = str(ctx.attr._relationships_bzl)
    lines = [
        "load({}, \"relationship\", \"relationships\")".format(_quote(bzl_label)),
        "",
    ]
    entries = []
    for index in range(len(ctx.attr.rule_kinds)):
        name = "declaration_{}".format(index)
        lines.extend([
            "relationship(",
            "    name = {},".format(_quote(name)),
            "    rule_module = {},".format(_quote(ctx.attr.rule_modules[index])),
            "    rule_kind = {},".format(_quote(ctx.attr.rule_kinds[index])),
            "    attribute = {},".format(_quote(ctx.attr.attributes[index])),
            "    requires_attributes = [{}],".format(", ".join([
                _quote(attribute)
                for attribute in ctx.attr.requires_attributes[index].split(" ")
                if attribute
            ])),
            "    relationship = {},".format(_quote(ctx.attr.relationships[index])),
            "    module_name = {},".format(_quote(ctx.attr.module_names[index])),
            ")",
            "",
        ])
        entries.append(":{}".format(name))
    lines.extend([
        "relationships(",
        "    name = \"relationships\",",
        "    visibility = [\"//visibility:public\"],",
        "    entries = [{}],".format(", ".join([_quote(entry) for entry in entries])),
        ")",
        "",
    ])
    ctx.file("BUILD.bazel", "\n".join(lines))

_relationships_repo = repository_rule(
    implementation = _relationships_repo_impl,
    attrs = {
        "rule_modules": attr.string_list(mandatory = True),
        "rule_kinds": attr.string_list(mandatory = True),
        "attributes": attr.string_list(mandatory = True),
        "requires_attributes": attr.string_list(mandatory = True),
        "relationships": attr.string_list(mandatory = True),
        "module_names": attr.string_list(mandatory = True),
        "_relationships_bzl": attr.label(
            allow_single_file = True,
            default = Label("//gather_metadata:relationships.bzl"),
        ),
    },
)

def merge_relationship_declarations(declarations):
    """Merge module declarations; root wins and non-root conflicts fail."""
    merged = {}
    conflicts = []
    for declaration in declarations:
        current = merged.get(declaration.key)
        if current == None:
            merged[declaration.key] = declaration
        elif current.relationship == declaration.relationship:
            if declaration.is_root and not current.is_root:
                merged[declaration.key] = declaration
        elif declaration.is_root and not current.is_root:
            merged[declaration.key] = declaration
        elif current.is_root and not declaration.is_root:
            continue
        else:
            conflicts.append(
                "conflicting relationship declarations for {}: module '{}' ({}) declares '{}' while module '{}' ({}) declares '{}'; the root module wins when it declares this key".format(
                    declaration.key,
                    current.module_name,
                    current.source,
                    current.relationship,
                    declaration.module_name,
                    declaration.source,
                    declaration.relationship,
                ),
            )
    return struct(
        declarations = [merged[key] for key in sorted(merged.keys())],
        conflicts = conflicts,
    )

def _relationship_extension_impl(module_ctx):
    declarations = []
    for module in module_ctx.modules:
        for tag in module.tags.declare:
            if tag.relationship not in _VALID_RELATIONSHIPS:
                fail(
                    "relationship declared by module '{}' must be one of {}; got '{}'".format(
                        module.name,
                        _VALID_RELATIONSHIPS,
                        tag.relationship,
                    ),
                )
            if not tag.rule_kind or not tag.attribute:
                fail("relationship rule_kind and attribute must be non-empty")
            if not module.is_root and tag.rule_module:
                fail("relationship rule_module may only be set by the root module")
            if module.is_root:
                rule_module = normalize_rule_module(tag.rule_module) if tag.rule_module else ""
            else:
                rule_module = module.name
            requires_attributes = sorted({attribute: True for attribute in tag.requires_attributes}.keys())
            key = "{}|{}|{}|{}".format(
                rule_module,
                tag.rule_kind,
                tag.attribute,
                ",".join(requires_attributes),
            )
            declarations.append(struct(
                key = key,
                rule_module = rule_module,
                rule_kind = tag.rule_kind,
                attribute = tag.attribute,
                requires_attributes = requires_attributes,
                relationship = tag.relationship,
                module_name = module.name,
                is_root = module.is_root,
                source = tag.name or "{}|{}".format(tag.rule_kind, tag.attribute),
            ))
    result = merge_relationship_declarations(declarations)
    if result.conflicts:
        fail(result.conflicts[0])
    _relationships_repo(
        name = "supply_chain_relationships",
        rule_modules = [d.rule_module for d in result.declarations],
        rule_kinds = [d.rule_kind for d in result.declarations],
        attributes = [d.attribute for d in result.declarations],
        requires_attributes = [" ".join(d.requires_attributes) for d in result.declarations],
        relationships = [d.relationship for d in result.declarations],
        module_names = [d.module_name for d in result.declarations],
    )

relationships = module_extension(
    implementation = _relationship_extension_impl,
    tag_classes = {
        "declare": tag_class(attrs = {
            "name": attr.string(),
            "rule_module": attr.string(),
            "rule_kind": attr.string(mandatory = True),
            "attribute": attr.string(mandatory = True),
            "requires_attributes": attr.string_list(),
            "relationship": attr.string(mandatory = True),
        }),
    },
)
