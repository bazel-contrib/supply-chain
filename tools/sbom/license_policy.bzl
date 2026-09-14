"""Bazel rule for evaluating a license coverage report."""

def _license_policy_impl(ctx):
    out = ctx.actions.declare_file(ctx.attr.out)
    ctx.actions.run(
        executable = ctx.executable._checker,
        arguments = [
            "--report",
            ctx.file.report.path,
            "--allowlist",
            ctx.file.allowlist.path,
            "--output",
            out.path,
        ],
        inputs = [ctx.file.report, ctx.file.allowlist],
        outputs = [out],
        mnemonic = "LicensePolicy",
    )
    return DefaultInfo(files = depset([out]))

license_policy = rule(
    implementation = _license_policy_impl,
    attrs = {
        "allowlist": attr.label(mandatory = True, allow_single_file = True),
        "out": attr.string(mandatory = True),
        "report": attr.label(mandatory = True, allow_single_file = True),
        "_checker": attr.label(
            default = "@supply-chain-go//cmd/license_policy",
            executable = True,
            cfg = "exec",
        ),
    },
)
