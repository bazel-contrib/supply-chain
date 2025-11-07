"""Declares rule `package_metadata_override`."""

load("@package_metadata//providers:package_metadata_info.bzl", "PackageMetadataInfo")
load("@package_metadata//providers:package_metadata_override_info.bzl", "PackageMetadataOverrideInfo")
load("//providers:package_specification_info.bzl", "PackageSpecificationInfo")
load("//rules:package_group.bzl", "package_group")

visibility("public")

def _package_metadata_override_rule_impl(ctx):
    return [
        PackageMetadataOverrideInfo(
            packages = ctx.attr.packages[PackageSpecificationInfo],
            metadata = ctx.attr.metadata[PackageMetadataInfo],
        ),
    ]

_package_metadata_override_rule = rule(
    implementation = _package_metadata_override_rule_impl,
    attrs = {
        "metadata": attr.label(
            mandatory = True,
            providers = [
                PackageMetadataInfo,
            ],
        ),
        "packages": attr.label(
            mandatory = True,
        ),
    },
    provides = [
        PackageMetadataOverrideInfo,
    ],
    doc = """
TODO
""".strip(),
)

def _parse_package_pattern(pattern):
    if pattern.endswith("..."):
        return pattern[:-3].rstrip("/"), None, True

    target_separator = pattern.rfind(":")
    if target_separator < 0:
        return pattern.rstrip("/"), None, False

    return pattern[:target_separator], pattern[(target_separator + 1):], False

def _package_metadata_override_impl(name, targets, metadata, visibility):
    patterns = {}
    for pattern in targets:
        if pattern.startswith("//") or pattern.startswith("-//"):
            label = "//:ignored"
            package, target, recursive = _parse_package_pattern(pattern.lstrip("-").lstrip("/"))
            config = {
                "negative": pattern.startswith("-"),
                "package": package,
                "recursive": recursive,
                "target": target,
            }
        elif pattern.startswith("@") or pattern.startswith("-@"):
            pattern_start = pattern.find("//")
            if pattern_start < 0:
                fail("'%s' is not a valid pattern".format(pattern))
                continue

            label = "{}//:ignored".format(pattern[:pattern_start].lstrip("-"))
            pkg, target, recursive = _parse_package_pattern(pattern.lstrip("-")[pattern_start:].lstrip("/"))
            config = {
                "negative": pattern.startswith("-"),
                "package": pkg,
                "recursive": recursive,
                "target": target,
            }
        else:
            fail("'%s' is not a valid pattern".format(pattern))
            continue

        patterns.setdefault(label, []).append(config)

    package_group(
        name = "{}_packages".format(name),
        packages = {label: json.encode(pattern) for label, pattern in patterns.items()},
    )

    _package_metadata_override_rule(
        name = name,
        visibility = visibility,
        packages = ":{}_packages".format(name),
        metadata = metadata,
    )

package_metadata_override = macro(
    implementation = _package_metadata_override_impl,
    attrs = {
        "metadata": attr.label(
            mandatory = True,
            configurable = False,
            doc = """
A `package_metadata` target to use for the provided packages.

This overrides any `package_metadata` directly declared by the packages.
""".strip(),
            providers = [
                PackageMetadataInfo,
            ],
        ),
        "targets": attr.string_list(
            mandatory = True,
            configurable = False,
            doc = """
A list of targets the override applies to.

This allows passing targets of the form

  - `//foo/...` matches all targets under `//foo` in the current repository.
  - `//foo` matches all targets in `//foo` (but not in subpackages).
  - `//foo:bar` matches exactly target `//foo:bar`.

This also supports referencing packages and targets in external repositories.

  - `@other_repo//foo/...` matches all targets under `//foo` in `other_repo`.
  - `@other_repo//foo` matches all targets in `//foo` in `other_repo` (but not
    in subpackages).
  - `@other_repo//foo:bar` matches exactly target `//foo:bar` in `other_repo`.
""".strip(),
        ),
    },
)
