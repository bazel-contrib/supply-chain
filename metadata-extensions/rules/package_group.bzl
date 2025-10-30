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

        repo = pattern["repo"]

        config.setdefault(repo, []).append(struct(
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

_package_group_rule = rule(
    implementation = _package_group_rule_impl,
    attrs = {
        "patterns": attr.string_list(
            mandatory = True,
        ),
    },
    provides = [
        PackageSpecificationInfo,
    ],
)

def _package_group_impl(name, packages, visibility):
    patterns = []
    for label, configs in packages.items():
        for config in json.decode(configs):
            patterns.append({
                "negative": config["negative"],
                "package": config["package"],
                "recursive": config["recursive"],
                "repo": label.repo_name,
                "target": config["target"],
            })

    _package_group_rule(
        name = name,
        visibility = visibility,
        patterns = [json.encode(pattern) for pattern in patterns],
    )

package_group = macro(
    implementation = _package_group_impl,
    attrs = {
        "packages": attr.label_keyed_string_dict(
            mandatory = True,
            configurable = False,
            doc = """
A dict from a dummy label to the parsed configuration.

Key: a label in the repository of the target pattern. Only `Label.repo_name` is
    used. Does not need to extist.
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
    doc = """
A starlark implementation of [package_group](https://bazel.build/reference/be/functions#package_group).

Unlike the native implementation in Bazel, this allows patterns referencing other repositories.

This has a very subtle API. Not for direct use.
""".strip(),
)
