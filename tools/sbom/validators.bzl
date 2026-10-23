"""Toolchain rules for upstream SBOM validators."""

def _cyclonedx_validator_toolchain_impl(ctx):
    return [
        platform_common.ToolchainInfo(
            files_to_run = ctx.attr.validator[DefaultInfo].files_to_run,
            validator = ctx.file.validator,
        ),
    ]

cyclonedx_validator_toolchain = rule(
    implementation = _cyclonedx_validator_toolchain_impl,
    attrs = {
        "validator": attr.label(
            allow_single_file = True,
            cfg = "exec",
            mandatory = True,
            doc = "CycloneDX CLI executable for the exec platform.",
        ),
    },
)

def _spdx_validator_toolchain_impl(ctx):
    return [
        platform_common.ToolchainInfo(
            files_to_run = ctx.attr.validator[DefaultInfo].files_to_run,
            validator = ctx.executable.validator,
        ),
    ]

spdx_validator_toolchain = rule(
    implementation = _spdx_validator_toolchain_impl,
    attrs = {
        "validator": attr.label(
            cfg = "exec",
            executable = True,
            mandatory = True,
            doc = "Executable SPDX tools-java validator wrapper for the exec platform.",
        ),
    },
)
