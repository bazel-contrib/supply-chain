load("sbom.bzl", "SbomProvider")

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
            ctx.attr.sbom[SbomProvider].config.path,
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
        "sbom": attr.label(),
        "format": attr.string(default = "json", values = ["json", "yaml", "tag-value"]),
        "out": attr.output(),
        "_spdx": attr.label(default = "@supply-chain-go//cmd/spdx"),
    }
)