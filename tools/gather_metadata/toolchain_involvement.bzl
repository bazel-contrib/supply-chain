"""Registry for classifying toolchain dependency edges."""

ToolchainInvolvementInfo = provider(
    doc = "Toolchain involvement overrides keyed by rule kind and normalized type.",
    fields = {
        "entries": "dict<string, string>: '<rule_kind>|<toolchain_type>' to involvement",
    },
)

_VALID_INVOLVEMENTS = ["runtime", "embedded", "build"]

def normalize_toolchain_type(label):
    """Return a toolchain label in apparent repository form."""
    value = str(label)
    if value.startswith("@@//"):
        return value[2:]
    if value.startswith("@@"):
        value = value[1:]
    if value.startswith("@") and not value.startswith("@@"):
        repository_end = value.find("//")
        if repository_end >= 0:
            repository = value[1:repository_end]
            for separator in ["+", "~"]:
                separator_index = repository.find(separator)
                if separator_index >= 0:
                    repository = repository[:separator_index]
            value = "@" + repository + value[repository_end:]
    return value

def _toolchain_involvement_impl(ctx):
    if ctx.attr.involvement not in _VALID_INVOLVEMENTS:
        fail("involvement must be one of {}".format(_VALID_INVOLVEMENTS))
    key = "{}|{}".format(ctx.attr.rule_kind, normalize_toolchain_type(ctx.attr.toolchain_type))
    return [ToolchainInvolvementInfo(entries = {key: ctx.attr.involvement})]

toolchain_involvement = rule(
    implementation = _toolchain_involvement_impl,
    attrs = {
        "rule_kind": attr.string(default = "*"),
        "toolchain_type": attr.string(mandatory = True),
        "involvement": attr.string(mandatory = True),
    },
)

def _toolchain_involvements_impl(ctx):
    entries = {}
    for dependency in ctx.attr.deps:
        for key, value in dependency[ToolchainInvolvementInfo].entries.items():
            entries[key] = value
    for entry in ctx.attr.entries:
        for key, value in entry[ToolchainInvolvementInfo].entries.items():
            entries[key] = value
    return [ToolchainInvolvementInfo(entries = entries)]

toolchain_involvements = rule(
    implementation = _toolchain_involvements_impl,
    attrs = {
        "entries": attr.label_list(providers = [ToolchainInvolvementInfo]),
        "deps": attr.label_list(providers = [ToolchainInvolvementInfo]),
    },
)

def toolchain_involvement_for(registry, rule_kind, toolchain_type):
    """Look up a specific rule override, then the wildcard rule override."""
    entries = registry.entries
    normalized_type = normalize_toolchain_type(toolchain_type)
    specific = "{}|{}".format(rule_kind, normalized_type)
    wildcard = "*|{}".format(normalized_type)
    if specific in entries:
        return entries[specific]
    return entries.get(wildcard)
