load("providers.bzl", "SbomInfo")

def _spdx_impl(ctx):
    out_path = ctx.attr.out.name if ctx.attr.out != None else "%s.txt" % ctx.attr.name
    out = ctx.actions.declare_file(out_path)
    inputs = depset(
        [],
        transitive = [
            ctx.attr._spdx[DefaultInfo].data_runfiles.files,
            ctx.attr.sbom[DefaultInfo].files,
        ],
    )
    ctx.actions.run(
        outputs=[out],
        inputs=inputs,
        executable=ctx.attr._spdx[DefaultInfo].files_to_run.executable,
        arguments=[
            "--config",
            ctx.attr.sbom[SbomInfo].config.path,
            "--out",
            out.path,
            "--format",
            ctx.attr.format,
        ],
    )
    return DefaultInfo(files=depset([out]))

spdx = rule(
    _spdx_impl,
    attrs = {
        "sbom": attr.label(doc = "The sbom target to generate the SPDX SBOM from."),
        "format": attr.string(default = "json", values = ["json", "yaml", "tag-value"], doc = "The output format for the SPDX SBOM."),
        "out": attr.output(doc = "The output file for the SPDX SBOM."),
        "_spdx": attr.label(default = "@supply-chain-go//cmd/spdx", doc = "The spdx tool to use."),
    }
)