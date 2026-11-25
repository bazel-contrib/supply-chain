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
    # same way as creationInfo.created of the SPDX reports. Bazel always writes
    # the time of the invocation to ctx.version_file (volatile-status.txt), also
    # without --stamp, so the file is only passed to the generator when the
    # build is stamped. Otherwise the generator writes the Unix epoch, which
    # keeps the output deterministic by default.
    if ctx.attr.stamp:
        extra_inputs.append(ctx.version_file)
        extra_args.extend(["--created_from_status_file", ctx.version_file.path])
    if ctx.file.serial_number:
        extra_inputs.append(ctx.file.serial_number)
        extra_args.extend(["--serial_number_file", ctx.file.serial_number.path])
    if strict:
        cyclonedx_validator = ctx.toolchains["//sbom:cyclonedx_validator_toolchain_type"]
        extra_args.extend(["--validator", cyclonedx_validator.validator.path])
        extra_tools.append(cyclonedx_validator.files_to_run)

    if not ctx.attr.embed_patch_files:
        extra_args.append("--embed_patch_files=false")

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

_cyclonedx = rule(
    _cyclonedx_impl,
    attrs = {
        "stamp": attr.bool(
            doc = "Whether the report is stamped with the time of the build, which is read from the workspace status. Set by the macro, which defaults it to whether the build is invoked with `--stamp`.",
        ),
        "sbom": attr.label(doc = "The sbom target to generate the CycloneDX SBOM from."),
        "format": attr.string(default = "json", values = ["json", "xml"], doc = "The output format for the CycloneDX SBOM."),
        "embed_patch_files": attr.bool(
            default = True,
            doc = """Whether to embed the content of patch files (the `file` of a `patch` attribute) in the report, as base64 in the pedigree of the patched package.

The embedded content makes the report self-contained but makes it larger, and discloses the patches to whoever receives the report. Set it to False to keep the patches local: they are still reported, with their type, the issues they resolve and the path of the file in the notes, but without their content. Patches given by `url` are never embedded.""",
        ),
        "out": attr.output(doc = "The output file for the CycloneDX SBOM."),
        "serial_number": attr.label(
            allow_single_file = True,
            doc = """Optional file holding the BOM serialNumber, in `urn:uuid:` form, unique to each generated BOM. The UUID must be in lowercase, as the CycloneDX schema requires. By default the BOM has no serialNumber.

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

def cyclonedx(name, stamp = None, **kwargs):
    """Generates a CycloneDX report from an `sbom` target.

    Args:
      name: A unique name for this target.
      stamp: Whether to stamp the report with the time of the build, in `metadata.timestamp`.
        By default it follows `--stamp`: the real time of the build with
        `--stamp`, and the Unix epoch otherwise, so that the report is
        deterministic. Bazel writes the time of the invocation to the workspace
        status also without `--stamp`, so the rule does not read it then.
      **kwargs: The other attributes of the rule: `sbom`, `format`, `out`
        and the optional ones described there.
    """
    _cyclonedx(
        name = name,
        stamp = select({
            Label("//sbom:stamp"): True,
            "//conditions:default": False,
        }) if stamp == None else stamp,
        **kwargs
    )
