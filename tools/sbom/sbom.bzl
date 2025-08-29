load("providers.bzl", "SbomInfo")
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
    transitive_inputs = []
    config = { "deps": [] }
    for m in transitive_metadata_info.metadata.to_list():
        config["deps"].append({
            "metadata": m.metadata.path
        })
        transitive_inputs.append(m.files)

    sbom_gen_config = ctx.actions.declare_file("{name}.sbom.config.json".format(name=ctx.attr.name))
    ctx.actions.write(sbom_gen_config, json.encode(config))

    return [
        DefaultInfo(files=depset(
            [sbom_gen_config],
            transitive=transitive_inputs
        )),
        SbomInfo(config=sbom_gen_config),
    ]

def sbom_rule(gathering_aspect):
    return rule(
        _sbom_impl,
        attrs = {
            "target": attr.label(aspects = [gathering_aspect], doc="The target for which to generate an SBOM."),
        },
    )

sbom = sbom_rule(gathering_aspect=gather_metadata_info)