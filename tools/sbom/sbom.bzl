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
    "@supply_chain_tools//gather_metadata:core.bzl",
    "gather_metadata_info_common",
    "should_traverse",
)
load(
    "@supply_chain_tools//gather_metadata:providers.bzl",
    "null_transitive_metadata_info",
    "TransitiveMetadataInfo",
)

ToolchainSbomInfo = provider()

def _gather_metadata_info_impl(target, ctx):
    return gather_metadata_info_common(
        target,
        ctx,
        want_providers = [PackageAttributeInfo, PackageMetadataInfo, LicenseKindInfo],
        provider_factory = TransitiveMetadataInfo,
        null_provider_instance = null_transitive_metadata_info,
        filter_func = should_traverse,
    )

gather_metadata_info = aspect(
    doc = """Collects metadata providers into a single TransitiveMetadataInfo provider.""",
    implementation = _gather_metadata_info_impl,
    attr_aspects = ["*"],
    attrs = {
        "_trace": attr.label(default = "@supply_chain_tools//gather_metadata:trace_target"),
    },
    provides = [TransitiveMetadataInfo],
    apply_to_generating_rules = True,
)

def _sbom_impl(ctx):
    transitive_metadata_info = ctx.attr.target[TransitiveMetadataInfo]
    sbom_toolchain = ctx.toolchains[SBOM_TOOLCHAIN_TYPE]
    out = ctx.actions.declare_file(ctx.attr.out.name if ctx.attr.out != None else "%s.txt" % ctx.attr.name)
    return sbom_toolchain.sbom_toolchain_info.generate_sbom(
        ctx,
        format=ctx.attr.format,
        out=out,
        info=transitive_metadata_info,
    )

SBOM_TOOLCHAIN_TYPE = "//sbom:sbom_toolchain_type"

def sbom_rule(gathering_aspect):
    return rule(
        _sbom_impl,
        attrs = {
            "target": attr.label(aspects = [gathering_aspect]),
            "out": attr.output(),
            "format": attr.string(),
        },
        toolchains = [
            SBOM_TOOLCHAIN_TYPE,
        ],
    )

sbom = sbom_rule(gathering_aspect=gather_metadata_info)