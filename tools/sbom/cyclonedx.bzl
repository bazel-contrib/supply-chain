load("@bazel_skylib//rules:common_settings.bzl", "BuildSettingInfo")
load("providers.bzl", "SbomInfo")

def _cyclonedx_impl(ctx):
    out_path = ctx.attr.out.name if ctx.attr.out != None else "%s.json" % ctx.attr.name
    out = ctx.actions.declare_file(out_path)
    inputs = depset(
        [],
        transitive = [
            ctx.attr._cyclonedx[DefaultInfo].data_runfiles.files,
            ctx.attr.sbom[DefaultInfo].files,
        ],
    )
    ctx.actions.run(
        outputs = [out],
        inputs = inputs,
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
        ],
    )

    # Validate the generated BOM against the vendored CycloneDX JSON Schema.
    # Disabled by default (see //sbom:validate_schemas); opt in with
    # --//sbom:validate_schemas, since re-validating against the full schema
    # on every build has a real cost. Only "json" has a JSON Schema to
    # validate against; "xml" output is not covered. See
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
                [out, ctx.file._cdx_schema] + ctx.files._cdx_schema_aux,
                transitive = [ctx.attr._schemavalidate[DefaultInfo].data_runfiles.files],
            ),
            executable = ctx.attr._schemavalidate[DefaultInfo].files_to_run,
            arguments = [
                "--schema",
                ctx.file._cdx_schema.path,
            ] + [arg for f in ctx.files._cdx_schema_aux for arg in ("--aux_schema", f.path)] + [
                "--instance",
                out.path,
                "--output",
                validation_out.path,
            ],
            mnemonic = "CycloneDxSchemaValidate",
        )
        output_groups["_validation"] = depset([validation_out])

    return [
        DefaultInfo(files = depset([out])),
        OutputGroupInfo(**output_groups),
    ]

cyclonedx = rule(
    _cyclonedx_impl,
    attrs = {
        "sbom": attr.label(doc = "The sbom target to generate the CycloneDX SBOM from."),
        "format": attr.string(default = "json", values = ["json", "xml"], doc = "The output format for the CycloneDX SBOM."),
        "out": attr.output(doc = "The output file for the CycloneDX SBOM."),
        "_cyclonedx": attr.label(default = "@supply-chain-go//cmd/cyclonedx", doc = "The cyclonedx tool to use.", executable = True, cfg = "exec"),
        "_schemavalidate": attr.label(default = "@supply-chain-go//cmd/schemavalidate", doc = "The JSON Schema validator to use.", executable = True, cfg = "exec"),
        "_cdx_schema": attr.label(
            default = "//sbom/schemas/cyclonedx:bom-1.6.schema.json",
            allow_single_file = True,
            doc = "The vendored root CycloneDX JSON Schema.",
        ),
        "_cdx_schema_aux": attr.label_list(
            default = [
                "//sbom/schemas/cyclonedx:spdx.schema.json",
                "//sbom/schemas/cyclonedx:jsf-0.82.schema.json",
            ],
            allow_files = True,
            doc = "Vendored schemas referenced by _cdx_schema via \"$ref\".",
        ),
        "_validate_schemas": attr.label(
            default = "//sbom:validate_schemas",
            doc = "Whether to run the schema-validation action. See //sbom:validate_schemas.",
        ),
    },
)
