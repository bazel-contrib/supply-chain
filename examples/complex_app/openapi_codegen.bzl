def _openapi_generate_sources_impl(ctx):
    output = ctx.outputs.out
    generator = ctx.attr.generator[DefaultInfo].files_to_run
    ctx.actions.run(
        executable = ctx.executable.generator,
        tools = [generator],
        inputs = [ctx.file.spec],
        outputs = [output],
        arguments = [
            ctx.file.spec.path,
            output.path,
        ],
        mnemonic = "OpenApiGenerateSources",
    )
    return [DefaultInfo(files = depset([output]))]

openapi_generate_sources = rule(
    implementation = _openapi_generate_sources_impl,
    attrs = {
        "generator": attr.label(
            executable = True,
            cfg = "exec",
            mandatory = True,
        ),
        "spec": attr.label(
            allow_single_file = True,
            mandatory = True,
        ),
        "out": attr.output(mandatory = True),
    },
    toolchains = ["@bazel_tools//tools/jdk:toolchain_type"],
)
