load("@supply_chain_tools//sbom:sbom.bzl", "ToolchainSbomInfo")

def _generate_sbom_factory(spdx_tool):
    def fn(ctx, out, format, info):
        if format == "spdx":
            sbom_tool=spdx_tool
        else:
            fail("format '{}' is not supported".format(format))
        transitive_inputs = []
        for m in info.metadata.to_list():
            transitive_inputs.append(m.files)
        ctx.actions.run(
            outputs=[out],
            inputs=depset(
                [],
                transitive=[
                    sbom_tool[DefaultInfo].default_runfiles.files,
                    sbom_tool[DefaultInfo].data_runfiles.files,
                ] + transitive_inputs
            ),
            executable=sbom_tool[DefaultInfo].files_to_run.executable,
            arguments=[
                out.path,
            ],
        )
        return DefaultInfo(files=depset([out]))
    return fn

def _sbom_toolchain_impl(ctx):
    return platform_common.ToolchainInfo(
        sbom_toolchain_info = ToolchainSbomInfo(
            generate_sbom = _generate_sbom_factory(
                spdx_tool=ctx.attr._spdx
            ),
        )
    )

sbom_toolchain = rule(
    _sbom_toolchain_impl,
    attrs = {
        "_spdx": attr.label(default = "//cmd/spdx", executable=True, cfg="exec"),
    },
)