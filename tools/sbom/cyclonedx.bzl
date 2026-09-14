load("providers.bzl", "SbomInfo")

def _cyclonedx_impl(ctx):
    out_path = ctx.attr.out.name if ctx.attr.out != None else "%s.json" % ctx.attr.name
    out = ctx.actions.declare_file(out_path)
    inputs = depset(
        [ctx.file.coverage] if ctx.file.coverage else [],
        transitive = [
            ctx.attr._cyclonedx[DefaultInfo].data_runfiles.files,
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
        executable = ctx.attr._cyclonedx[DefaultInfo].files_to_run,
        arguments = arguments,
    )
    return DefaultInfo(files = depset([out]))

cyclonedx = rule(
    _cyclonedx_impl,
    attrs = {
        "sbom": attr.label(doc = "The sbom target to generate the CycloneDX SBOM from."),
        "coverage": attr.label(allow_single_file = True, doc = "Optional license coverage report to embed in the CycloneDX component fields."),
        "format": attr.string(default = "json", values = ["json", "xml"], doc = "The output format for the CycloneDX SBOM."),
        "out": attr.output(doc = "The output file for the CycloneDX SBOM."),
        "_cyclonedx": attr.label(default = "@supply-chain-go//cmd/cyclonedx", doc = "The cyclonedx tool to use.", executable = True, cfg = "exec"),
    },
)
