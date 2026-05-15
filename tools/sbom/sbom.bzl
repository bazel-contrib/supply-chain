load("providers.bzl", "SbomInfo")
load(
    "@supply_chain_tools//gather_metadata:gather_metadata.bzl",
    "gather_metadata_info",
)
load(
    "@supply_chain_tools//gather_metadata:serialization.bzl",
    "metadata_info_to_json",
)
load(
    "@supply_chain_tools//gather_metadata:providers.bzl",
    "TransitiveMetadataInfo",
)

def _sbom_impl(ctx):
    transitive_metadata_info = ctx.attr.target[TransitiveMetadataInfo]

    # Get graph-only format from gather_metadata
    json_strings = metadata_info_to_json(transitive_metadata_info)
    graph_json_content = json_strings[0] if json_strings else "{}"

    # Write graph JSON
    graph_json = ctx.actions.declare_file("{name}.graph.json".format(name=ctx.attr.name))
    ctx.actions.write(graph_json, graph_json_content)

    # Run cmd/sbom action to compute classifications
    classifications_json = ctx.actions.declare_file("{name}.sbom-classifications.json".format(name=ctx.attr.name))

    args = [
        "--input",
        graph_json.path,
        "--output",
        classifications_json.path,
    ]

    if not ctx.attr.require_root_metadata:
        args.append("--allow_missing_root_metadata")

    ctx.actions.run(
        outputs = [classifications_json],
        inputs = [graph_json],
        executable = ctx.attr._sbom_cmd[DefaultInfo].files_to_run,
        arguments = args,
        mnemonic = "SbomClassify",
    )

    # Collect all metadata files as transitive inputs
    transitive_inputs = []
    for transitive_metadata in transitive_metadata_info.transitive.to_list():
        for m in transitive_metadata.metadata.to_list():
            if hasattr(m, "files"):
                transitive_inputs.append(m.files)

    return [
        DefaultInfo(files = depset(
            [graph_json, classifications_json],
            transitive = transitive_inputs,
        )),
        SbomInfo(
            graph = graph_json,
            classifications = classifications_json,
        ),
    ]

def sbom_rule(gathering_aspect):
    return rule(
        _sbom_impl,
        attrs = {
            "target": attr.label(aspects = [gathering_aspect], doc = "The target for which to generate an SBOM."),
            "require_root_metadata": attr.bool(
                default = True,
                doc = "If True, fail if the root target has no package metadata. If False, generate an SBOM without a root component.",
            ),
            "_sbom_cmd": attr.label(
                default = "@supply-chain-go//cmd/sbom",
                executable = True,
                cfg = "exec",
            ),
        },
    )

sbom = sbom_rule(gathering_aspect=gather_metadata_info)
