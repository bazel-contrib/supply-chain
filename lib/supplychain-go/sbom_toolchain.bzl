load("@supply_chain_tools//sbom:sbom.bzl", "ToolchainSbomInfo")

def _generate_spdx(ctx, out, info):
    print(ctx)
    print(out)
    print(info)
    fail("very good")

def _generate_sbom(ctx, out, format, info):
    if format == "spdx":
        return _generate_spdx(ctx, out, info)
    fail("format '{}' is not supported".format(format))

def _sbom_toolchain_impl(ctx):
    return platform_common.ToolchainInfo(
        sbom_toolchain_info = ToolchainSbomInfo(
            generate_sbom = _generate_sbom,
        )
    )

sbom_toolchain = rule(
    _sbom_toolchain_impl,
    attrs = {
        "_spdx": attr.label(default = "//cmd/spdx", executable=True, cfg="exec"),
    },
)