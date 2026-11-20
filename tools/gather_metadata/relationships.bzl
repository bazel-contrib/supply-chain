"""Registry for classifying dependency edges by rule module, kind, and attribute."""

RelationshipInfo = provider(
    doc = "Dependency relationship overrides keyed by module, rule kind, attribute, and required attributes.",
    fields = {"entries": "dict<string, struct>: dependency relationship declarations"},
)

_VALID_RELATIONSHIPS = [
    "depends_on",
    "contains",
    "static_link",
    "dynamic_link",
    "test_dependency",
    "dev_dependency",
]

def normalize_rule_module(module_name):
    """Normalize a canonical Bazel repository name to its Bzlmod module name."""
    for separator in ["+", "~"]:
        separator_index = module_name.find(separator)
        if separator_index >= 0:
            return module_name[:separator_index]
    return module_name

def _entry_key(rule_module, rule_kind, attribute, requires_attributes):
    return "{}|{}|{}|{}".format(
        rule_module,
        rule_kind,
        attribute,
        ",".join(sorted(requires_attributes)),
    )

def _relationship_impl(ctx):
    if ctx.attr.relationship not in _VALID_RELATIONSHIPS:
        fail("relationship must be one of {}; got '{}'".format(_VALID_RELATIONSHIPS, ctx.attr.relationship))
    key = _entry_key(
        ctx.attr.rule_module,
        ctx.attr.rule_kind,
        ctx.attr.attribute,
        ctx.attr.requires_attributes,
    )
    return [RelationshipInfo(entries = {
        key: struct(
            rule_module = ctx.attr.rule_module,
            rule_kind = ctx.attr.rule_kind,
            attribute = ctx.attr.attribute,
            requires_attributes = tuple(sorted(ctx.attr.requires_attributes)),
            relationship = ctx.attr.relationship,
            module_name = ctx.attr.module_name,
        ),
    })]

relationship = rule(
    implementation = _relationship_impl,
    attrs = {
        "rule_module": attr.string(),
        "rule_kind": attr.string(mandatory = True),
        "attribute": attr.string(mandatory = True),
        "requires_attributes": attr.string_list(),
        "relationship": attr.string(mandatory = True),
        "module_name": attr.string(mandatory = True),
    },
)

def relationships_for(registry, rule_module, rule_kind, attribute, rule_attributes):
    """Resolve the most-specific matching relationship, failing on ambiguity."""
    rule_module = normalize_rule_module(rule_module)
    available_attributes = {name: True for name in rule_attributes}
    matches = []
    for key, entry in registry.entries.items():
        if entry.rule_kind not in [rule_kind, "*"] or entry.attribute != attribute:
            continue
        if entry.rule_module and entry.rule_module != rule_module:
            continue
        if any([required not in available_attributes for required in entry.requires_attributes]):
            continue

        module_specificity = 1 if entry.rule_module else 0
        kind_specificity = 1 if entry.rule_kind == rule_kind else 0
        matches.append((
            (module_specificity, kind_specificity, len(entry.requires_attributes)),
            key,
            entry,
        ))

    if not matches:
        return "depends_on"

    best_specificity = max([match[0] for match in matches])
    best_matches = [match for match in matches if match[0] == best_specificity]
    best_matches = sorted(best_matches, key = lambda match: match[1])
    selected = best_matches[0][2]
    for match in best_matches[1:]:
        other = match[2]
        if selected.relationship != other.relationship:
            fail(
                "ambiguous relationships for rule kind '{}' attribute '{}': '{}' declared by module '{}' and '{}' declared by module '{}' both match (rule module '{}')".format(
                    rule_kind,
                    attribute,
                    selected.relationship,
                    selected.module_name,
                    other.relationship,
                    other.module_name,
                    rule_module or "<unmarked>",
                ),
            )
    return selected.relationship

def _relationships_impl(ctx):
    entries = {}
    for dependency in ctx.attr.entries:
        for key, value in dependency[RelationshipInfo].entries.items():
            entries[key] = value
    return [RelationshipInfo(entries = entries)]

relationships = rule(
    implementation = _relationships_impl,
    attrs = {"entries": attr.label_list(providers = [RelationshipInfo])},
)
