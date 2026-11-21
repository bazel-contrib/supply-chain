"""Declares rule `package_group`."""

load("//providers:package_specification_info.bzl", "PackageSpecificationInfo")

visibility([
    "//...",
])

def _package_group_rule_impl(ctx):
    positive = {}
    negative = {}
    for pattern in [json.decode(pattern) for pattern in ctx.attr.patterns]:
        if pattern["negative"]:
            config = negative
        else:
            config = positive

        config.setdefault(pattern["repo"], []).append(struct(
            package = pattern["package"],
            target = pattern["target"],
            recursive = pattern["recursive"],
        ))

    return [
        PackageSpecificationInfo(
            negative = negative,
            positive = positive,
        ),
    ]

def package_group(name, packages, visibility = None):
    """Creates a package specification from repository-qualified patterns.

    Unlike Bazel's native package_group, this supports patterns in other repos.
    """
    patterns = []
    for label, configs in packages.items():
        repo = Label(label).repo_name if label.startswith("@") else ""
        for config in json.decode(configs):
            patterns.append({
                "negative": config["negative"],
                "package": config["package"],
                "recursive": config["recursive"],
                "repo": repo,
                "target": config["target"],
            })

    _package_group_rule(
        name = name,
        visibility = visibility,
        patterns = [json.encode(pattern) for pattern in patterns],
    )

_package_group_rule = rule(
    implementation = _package_group_rule_impl,
    attrs = {
        "patterns": attr.string_list(
            mandatory = True,
            doc = """
Repository-qualified patterns and their configuration data, encoded as JSON.
Value: a `json.encode()`'d list of dicts defining the configuration of the
       target pattern.

       ```starlark
       {
          "negative": True | False
          "pattern": "path/to/package/...",
       }
       ```
""".strip(),
        ),
    },
    provides = [
        PackageSpecificationInfo,
    ],
)
