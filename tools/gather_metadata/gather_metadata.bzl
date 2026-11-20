"""Rules and macros for collecting metadata providers."""

load("@bazel_features//:features.bzl", "bazel_features")
load(
    "@package_metadata//:defs.bzl",
    "PackageAttributeInfo",
    "PackageMetadataInfo",
)
load(
    "@package_metadata//licenses/providers:license_kind_info.bzl",
    "LicenseKindInfo",
)
load(
    ":core.bzl",
    "gather_metadata_info_common",
    "should_traverse",
)
load(":providers.bzl", "TransitiveMetadataInfo", "null_transitive_metadata_info")
load(
    ":serialization.bzl",
    "metadata_info_to_json",
    "write_metadata_info",
)

_gather_metadata_attrs = {
    "_include_build_tools": attr.label(
        default = "//sbom:include_build_tools",
        doc = "Whether to retain and classify exec-configuration dependencies.",
    ),
    "_toolchain_involvements": attr.label(
        default = "@supply_chain_toolchain_involvements//:toolchain_involvements",
        doc = "Toolchain involvement registry generated from module declarations.",
    ),
    "_relationships": attr.label(
        default = "@supply_chain_relationships//:relationships",
        doc = "Dependency relationship registry generated from module declarations.",
    ),
}

def _gather_metadata_info_impl(target, ctx):
    return gather_metadata_info_common(
        target,
        ctx,
        want_providers = [PackageAttributeInfo, PackageMetadataInfo, LicenseKindInfo],
        provider_factory = TransitiveMetadataInfo,
        null_provider_instance = null_transitive_metadata_info,
        filter_func = should_traverse,
        enable_intermediate_pruning = False,
    )

def _create_gather_metadata_info_aspect():
    args = {
        "doc": "Collects metadata providers into a single TransitiveMetadataInfo provider.",
        "implementation": _gather_metadata_info_impl,
        "attr_aspects": ["*"],
        "attrs": _gather_metadata_attrs,
        "provides": [TransitiveMetadataInfo],
        "apply_to_generating_rules": True,
    }
    if bazel_features.rules.supports_toolchains_aspects_star:
        args["toolchains_aspects"] = ["*"]
    return aspect(**args)

gather_metadata_info = _create_gather_metadata_info_aspect()

def _write_metadata_info_impl(target, ctx):
    """Write transitive metadata info into a JSON file.

    Args:
      target: The target of the aspect.
      ctx: The aspect evaluation context.

    Returns:
      OutputGroupInfo with metadata files
    """

    if not TransitiveMetadataInfo in target:
        return [OutputGroupInfo(licenses = depset())]
    info = target[TransitiveMetadataInfo]

    # Write the output file for the target
    name = "%s_metadata_info.json" % ctx.label.name
    json_strings = metadata_info_to_json(info)
    content = json_strings[0] if json_strings else "{}"
    out = ctx.actions.declare_file(name)
    ctx.actions.write(
        output = out,
        content = content,
    )

    return [OutputGroupInfo(metadata = depset([out]))]

gather_metadata_info_and_write = aspect(
    doc = """Collects TransitiveMetadataInfo providers and writes JSON representation to a file.

    Usage:
      bazel build //some:target \
          --aspects=@supply_chain_tools//gather_metadata:gather_metadata.bzl%gather_metadata_info_and_write \
          --output_groups=metadata

    Output format:
      Graph-only JSON with the structure:
      {
        "schema_version": "1.0",
        "root_target": "//some:target",
        "nodes": [
          {
            "label": "//some:target",
            "metadata_file": "bazel-bin/some/target.package-metadata.json"
          }
        ],
        "edges": [
          {
            "from": "//some:target",
            "to": "//dep:target",
            "type": "depends_on",
            "involvement": "build",
            "relationship": "contains"
          }
        ]
      }

    The optional edge "involvement" is "build" or "embedded". Runtime edges
    omit the field to preserve the original graph format. "relationship" is
    omitted for the default depends_on relationship.
    """,
    implementation = _write_metadata_info_impl,
    attr_aspects = ["*"],
    provides = [OutputGroupInfo],
    requires = [gather_metadata_info],
    apply_to_generating_rules = True,
)
