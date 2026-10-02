load("@bazel_skylib//rules:common_settings.bzl", "BuildSettingInfo")
load("providers.bzl", "SbomInfo")

def _spdx_impl(ctx):
    out_path = (
        ctx.attr.out.name if ctx.attr.out != None else "%s.txt" % ctx.attr.name if ctx.attr.format == "tag-value" else "%s.json" % ctx.attr.name if ctx.attr.format == "json" else "%s.yaml" % ctx.attr.name
    )

    out = ctx.actions.declare_file(out_path)

    # creationInfo.created (SPDX 2.3, mandatory) comes from Bazel's own
    # build-stamping mechanism: ctx.version_file (volatile-status.txt) holds
    # a real BUILD_TIMESTAMP only when the build is invoked with --stamp;
    # otherwise Bazel substitutes a fixed placeholder, which keeps this
    # action's output (and therefore its cache key) deterministic by
    # default. See https://bazel.build/reference/be/make-variables#stamp-flag.
    extra_inputs = []
    extra_args = []
    if ctx.version_file:
        extra_inputs.append(ctx.version_file)
        extra_args.extend(["--created_from_status_file", ctx.version_file.path])

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

    # Validate the generated document against the vendored SPDX 2.3 JSON
    # Schema. Disabled by default (see //sbom:validate_schemas); opt in with
    # --//sbom:validate_schemas, since re-validating against the full schema
    # on every build has a real cost. Only "json" has a JSON Schema to
    # validate against; "yaml" and "tag-value" are not covered. See
    # https://bazel.build/extending/rules#validation-actions: this output is
    # never part of DefaultInfo, only of the "_validation" output group, so
    # it runs (when enabled) without gating anything that depends on this
    # target.
    output_groups = {}
    if ctx.attr.format == "json" and ctx.attr._validate_schemas[BuildSettingInfo].value:
        validation_out = ctx.actions.declare_file("%s.schema_valid" % ctx.attr.name)
        ctx.actions.run(
            outputs = [validation_out],
            inputs = depset(
                [out, ctx.file._spdx_schema],
                transitive = [ctx.attr._schemavalidate[DefaultInfo].data_runfiles.files],
            ),
            executable = ctx.attr._schemavalidate[DefaultInfo].files_to_run,
            arguments = [
                "--schema",
                ctx.file._spdx_schema.path,
                "--instance",
                out.path,
                "--output",
                validation_out.path,
            ],
            mnemonic = "SpdxSchemaValidate",
        )
        output_groups["_validation"] = depset([validation_out])

    return [
        DefaultInfo(files = depset([out])),
        OutputGroupInfo(**output_groups),
    ]

spdx = rule(
    _spdx_impl,
    attrs = {
        "sbom": attr.label(doc = "The sbom target to generate the SPDX SBOM from."),
        "format": attr.string(default = "json", values = ["json", "yaml", "tag-value"], doc = "The output format for the SPDX SBOM."),
        "out": attr.output(doc = "The output file for the SPDX SBOM."),
        "_spdx": attr.label(default = "@supply-chain-go//cmd/spdx", doc = "The spdx tool to use.", executable = True, cfg = "exec"),
        "_schemavalidate": attr.label(default = "@supply-chain-go//cmd/schemavalidate", doc = "The JSON Schema validator to use.", executable = True, cfg = "exec"),
        "_spdx_schema": attr.label(
            default = "//sbom/schemas/spdx:spdx-schema.json",
            allow_single_file = True,
            doc = "The vendored SPDX 2.3 JSON Schema.",
        ),
        "_validate_schemas": attr.label(
            default = "//sbom:validate_schemas",
            doc = "Whether to run the schema-validation action. See //sbom:validate_schemas.",
        ),
    },
)
