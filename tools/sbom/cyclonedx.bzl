load("@bazel_skylib//rules:common_settings.bzl", "BuildSettingInfo")
load("providers.bzl", "SbomInfo")

def _cyclonedx_impl(ctx):
    out_path = ctx.attr.out.name if ctx.attr.out != None else "%s.json" % ctx.attr.name
    out = ctx.actions.declare_file(out_path)
    strict = ctx.attr.format in ["json", "xml"] and ctx.attr._strict_validations[BuildSettingInfo].value
    extra_args = []
    extra_tools = []
    if strict:
        cyclonedx_validator = ctx.toolchains["//sbom:cyclonedx_validator_toolchain_type"]
        extra_args.extend(["--validator", cyclonedx_validator.validator.path])
        extra_tools.append(cyclonedx_validator.files_to_run)

    inputs = depset(
        transitive = [
            ctx.attr._cyclonedx[DefaultInfo].data_runfiles.files,
            ctx.attr.sbom[DefaultInfo].files,
        ],
    )
    ctx.actions.run(
        outputs = [out],
        inputs = inputs,
        tools = extra_tools,
        executable = ctx.attr._cyclonedx[DefaultInfo].files_to_run,
        arguments = [
            "--graph",
            ctx.attr.sbom[SbomInfo].graph.path,
            "--classifications",
            ctx.attr.sbom[SbomInfo].classifications.path,
            "--out",
            out.path,
            "--format",
            ctx.attr.format,
        ] + extra_args,
    )

    return [
        DefaultInfo(files = depset([out])),
    ]

cyclonedx = rule(
    _cyclonedx_impl,
    attrs = {
        "sbom": attr.label(doc = "The sbom target to generate the CycloneDX SBOM from."),
        "format": attr.string(default = "json", values = ["json", "xml"], doc = "The output format for the CycloneDX SBOM."),
        "out": attr.output(doc = "The output file for the CycloneDX SBOM."),
        "_cyclonedx": attr.label(default = "@supply-chain-go//cmd/cyclonedx", doc = "The cyclonedx tool to use.", executable = True, cfg = "exec"),
        "_strict_validations": attr.label(
            default = "//sbom:strict_validations",
            doc = "Whether to validate generated reports with the upstream CycloneDX CLI. See //sbom:strict_validations.",
        ),
    },
    toolchains = ["//sbom:cyclonedx_validator_toolchain_type"],
)
