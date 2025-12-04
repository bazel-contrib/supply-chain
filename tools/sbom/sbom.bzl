load("providers.bzl", "SbomInfo")
load(
    "@supply_chain_tools//gather_metadata:gather_metadata.bzl",
    "gather_metadata_info",
)
load(
    "@supply_chain_tools//gather_metadata:providers.bzl",
    "TransitiveMetadataInfo",
)

def _sbom_impl(ctx):
    transitive_metadata_info = ctx.attr.target[TransitiveMetadataInfo]
    transitive_inputs = []
    config = { "deps": [] }
    for transitive_metadata in transitive_metadata_info.trans.to_list():
        for m in transitive_metadata.metadata.to_list():
            if hasattr(m, "metadata"):
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
