load("providers.bzl", "SbomInfo")

def _cdx_impl(ctx):
    out_path = ctx.attr.out.name if ctx.attr.out != None else "%s.json" % ctx.attr.name
    out = ctx.actions.declare_file(out_path)
    inputs = depset(
        [],
        transitive = [
            ctx.attr._cdx[DefaultInfo].data_runfiles.files,
            ctx.attr.sbom[DefaultInfo].files,
        ],
    )
    ctx.actions.run(
        outputs = [out],
        inputs = inputs,
        executable = ctx.attr._cdx[DefaultInfo].files_to_run.executable,
        arguments = [
            "--config",
            ctx.attr.sbom[SbomInfo].config.path,
            "--out",
            out.path,
            "--format",
            ctx.attr.format,
        ],
    )
    return DefaultInfo(files = depset([out]))

cdx = rule(
    _cdx_impl,
    attrs = {
        "sbom": attr.label(doc = "The sbom target to generate the CycloneDX SBOM from."),
        "format": attr.string(default = "json", values = ["json", "xml"], doc = "The output format for the CycloneDX SBOM."),
        "out": attr.output(doc = "The output file for the CycloneDX SBOM."),
        "_cdx": attr.label(default = "//tools/sbom/cmd/cdx:cdx", doc = "The cdx tool to use."),
    },
)
