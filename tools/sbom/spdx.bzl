load("@bazel_skylib//rules:common_settings.bzl", "BuildSettingInfo")
load("providers.bzl", "SbomInfo")

def _spdx_impl(ctx):
    out_path = (
        ctx.attr.out.name if ctx.attr.out != None else "%s.txt" % ctx.attr.name if ctx.attr.format == "tag-value" else "%s.yaml" % ctx.attr.name if ctx.attr.format == "yaml" else "%s.json" % ctx.attr.name
    )

    out = ctx.actions.declare_file(out_path)
    strict = ctx.attr._strict_validations[BuildSettingInfo].value

    # creationInfo.created (mandatory in SPDX 2.3 and 3.0.1) comes from Bazel's own
    # build-stamping mechanism: ctx.version_file (volatile-status.txt) holds
    # a real BUILD_TIMESTAMP only when the build is invoked with --stamp;
    # otherwise Bazel substitutes a fixed placeholder, which keeps this
    # action's output (and therefore its cache key) deterministic by
    # default. See https://bazel.build/reference/be/make-variables#stamp-flag.
    extra_inputs = []
    extra_args = []
    extra_tools = []
    if ctx.version_file:
        extra_inputs.append(ctx.version_file)
        extra_args.extend(["--created_from_status_file", ctx.version_file.path])
    if ctx.file.document_namespace:
        extra_inputs.append(ctx.file.document_namespace)
        extra_args.extend(["--document_namespace_file", ctx.file.document_namespace.path])
    if strict:
        spdx_validator = ctx.toolchains["//sbom:spdx_validator_toolchain_type"]
        extra_args.extend(["--validator", spdx_validator.validator.path])
        extra_tools.append(spdx_validator.files_to_run)

    inputs = depset(
        extra_inputs,
        transitive = [
            ctx.attr._spdx[DefaultInfo].data_runfiles.files,
            ctx.attr.sbom[DefaultInfo].files,
        ],
    )
    ctx.actions.run(
        outputs = [out],
        inputs = inputs,
        tools = extra_tools,
        executable = ctx.attr._spdx[DefaultInfo].files_to_run,
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

spdx = rule(
    _spdx_impl,
    attrs = {
        "sbom": attr.label(doc = "The sbom target to generate the SPDX SBOM from."),
        "format": attr.string(
            default = "json",
            values = ["json", "yaml", "tag-value", "json-ld"],
            doc = "The output format for the SPDX SBOM. `json`, `yaml` and `tag-value` write SPDX 2.3; `json-ld` writes SPDX 3.0.1.",
        ),
        "out": attr.output(doc = "The output file for the SPDX SBOM."),
        "document_namespace": attr.label(
            allow_single_file = True,
            doc = """Optional file holding the SPDX document namespace: an absolute URI without a fragment, unique to each generated document. By default the namespace is derived from the purl of the SBOM subject, which is the same for every build of it.

    This is a file, not a string, because the value that makes a document's
    namespace unique is normally produced by the build, for example from its
    workspace status with `expand_template` and `stamp_substitutions`. A string
    attribute is fixed when the BUILD file is loaded, so it could only hold the
    same value for every build. A file is only read when the action runs, which
    also keeps the report cacheable when the file does not change.""",
        ),
        "_spdx": attr.label(default = "@supply-chain-go//cmd/spdx", doc = "The spdx tool to use.", executable = True, cfg = "exec"),
        "_strict_validations": attr.label(
            default = "//sbom:strict_validations",
            doc = "Whether to validate generated reports with upstream SPDX tools-java. See //sbom:strict_validations.",
        ),
    },
    toolchains = ["//sbom:spdx_validator_toolchain_type"],
)
