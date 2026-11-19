load("@bazel_skylib//rules:common_settings.bzl", "BuildSettingInfo")
load("providers.bzl", "SbomInfo")

def _cyclonedx_impl(ctx):
    out_path = ctx.attr.out.name if ctx.attr.out != None else "%s.json" % ctx.attr.name
    out = ctx.actions.declare_file(out_path)
    strict = ctx.attr.format in ["json", "xml"] and ctx.attr._strict_validations[BuildSettingInfo].value
    extra_args = []
    extra_tools = []
    extra_inputs = []

    # metadata.timestamp comes from Bazel's own build-stamping mechanism, in the
    # same way as creationInfo.created of the SPDX reports: ctx.version_file
    # (volatile-status.txt) holds a real BUILD_TIMESTAMP only when the build is
    # invoked with --stamp, and a fixed placeholder otherwise, which keeps the
    # output deterministic by default.
    if ctx.version_file:
        extra_inputs.append(ctx.version_file)
        extra_args.extend(["--created_from_status_file", ctx.version_file.path])
    if ctx.file.serial_number:
        extra_inputs.append(ctx.file.serial_number)
        extra_args.extend(["--serial_number_file", ctx.file.serial_number.path])
    if strict:
        cyclonedx_validator = ctx.toolchains["//sbom:cyclonedx_validator_toolchain_type"]
        extra_args.extend(["--validator", cyclonedx_validator.validator.path])
        extra_tools.append(cyclonedx_validator.files_to_run)

    inputs = depset(
        extra_inputs,
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
        "serial_number": attr.label(
            allow_single_file = True,
            doc = """Optional file holding the BOM serialNumber, in `urn:uuid:` form, unique to each generated BOM. By default the BOM has no serialNumber.

    This is a file, not a string, because the value that makes a BOM unique is
    normally produced by the build, for example from its workspace status with
    `expand_template` and `stamp_substitutions`. A string attribute is fixed
    when the BUILD file is loaded, so it could only hold the same value for
    every build. A file is only read when the action runs, which also keeps
    the report cacheable when the file does not change.""",
        ),
        "_cyclonedx": attr.label(default = "@supply-chain-go//cmd/cyclonedx", doc = "The cyclonedx tool to use.", executable = True, cfg = "exec"),
        "_strict_validations": attr.label(
            default = "//sbom:strict_validations",
            doc = "Whether to validate generated reports with the upstream CycloneDX CLI. See //sbom:strict_validations.",
        ),
    },
    toolchains = ["//sbom:cyclonedx_validator_toolchain_type"],
)
