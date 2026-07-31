"""Bazel rule for producing a license metadata coverage report."""

load(":providers.bzl", "SbomInfo")

def _license_coverage_impl(ctx):
    out = ctx.actions.declare_file(ctx.attr.out)
    args = [
        "--graph",
        ctx.attr.sbom[SbomInfo].graph.path,
        "--output",
        out.path,
    ]
    inputs = ctx.attr.sbom[DefaultInfo].files.to_list()
    if ctx.file.resolutions:
        args.extend(["--resolutions", ctx.file.resolutions.path])
        inputs.append(ctx.file.resolutions)
    ctx.actions.run(
        executable = ctx.executable._reporter,
        arguments = args,
        inputs = inputs,
        outputs = [out],
        mnemonic = "LicenseCoverage",
    )
    return DefaultInfo(files = depset([out]))

license_coverage = rule(
    implementation = _license_coverage_impl,
    attrs = {
        "out": attr.string(mandatory = True),
        "resolutions": attr.label(allow_single_file = True),
        "sbom": attr.label(mandatory = True, providers = [SbomInfo]),
        "_reporter": attr.label(
            default = "@supply-chain-go//cmd/license_coverage",
            executable = True,
            cfg = "exec",
        ),
    },
)
