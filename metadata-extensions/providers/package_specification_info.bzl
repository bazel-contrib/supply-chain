"""Declares provider `PackageSpecificationInfo`."""

visibility([
    "//...",
])

def _contains_impl(configs, label):
    config = configs.get(label.repo_name, None)
    if not config:
        return False

    for cfg in config:
        if cfg.recursive:
            if label.package.startswith(cfg.package):
                return True
            continue

        if label.package != cfg.package:
            continue
        
        if cfg.target != None:
            if label.name == cfg.target:
                return True
        else:
            return True

    return False

def _contains(negative, positive, label):
    if _contains_impl(negative, label):
        return False

    return _contains_impl(positive, label)

def _init(negative, positive):
    return {
        "contains": lambda target: _contains(negative, positive, target),
    }

PackageSpecificationInfo, _create = provider(
    doc = """
Information about transitive package specifications used in package groups.

This is a Starlark implementation of [PackageSpecificationInfo](https://bazel.build/rules/lib/providers/PackageSpecificationInfo).
""".strip(),
    fields = {
        "contains": """
Checks if a target exists in a package group.

Parameter **MUST** be a label.
""".strip(),
    },
    init = _init,
)
