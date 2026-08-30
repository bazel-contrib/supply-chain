load("providers.bzl", "SbomInfo")

def _spdx_impl(ctx):
    out_path = (
        ctx.attr.out.name if ctx.attr.out != None else
        "%s.txt" % ctx.attr.name if ctx.attr.format == "tag-value" else
        "%s.json" % ctx.attr.name if ctx.attr.format == "json" else
        "%s.yaml" % ctx.attr.name
    )

    out = ctx.actions.declare_file(out_path)
    inputs = depset(
        [ctx.file.coverage] if ctx.file.coverage else [],
        transitive = [
            ctx.attr._spdx[DefaultInfo].data_runfiles.files,
            ctx.attr.sbom[DefaultInfo].files,
        ],
    )
    arguments = [
        "--graph",
        ctx.attr.sbom[SbomInfo].graph.path,
        "--classifications",
        ctx.attr.sbom[SbomInfo].classifications.path,
        "--out",
        out.path,
        "--format",
        ctx.attr.format,
    ]
    if ctx.file.coverage:
        arguments.extend(["--coverage", ctx.file.coverage.path])
    ctx.actions.run(
        outputs = [out],
        inputs = inputs,
        executable = ctx.attr._spdx[DefaultInfo].files_to_run,
        arguments = arguments,
    )
    return DefaultInfo(files = depset([out]))

spdx = rule(
    _spdx_impl,
    attrs = {
        "sbom": attr.label(doc = "The sbom target to generate the SPDX SBOM from."),
        "coverage": attr.label(allow_single_file = True, doc = "Optional license coverage report to embed in the SPDX package fields."),
        "format": attr.string(default = "json", values = ["json", "yaml", "tag-value"], doc = "The output format for the SPDX SBOM."),
        "out": attr.output(doc = "The output file for the SPDX SBOM."),
        "_spdx": attr.label(default = "@supply-chain-go//cmd/spdx", doc = "The spdx tool to use.", executable = True, cfg = "exec"),
    },
)
