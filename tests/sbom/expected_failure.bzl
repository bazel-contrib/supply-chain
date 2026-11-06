"""Checks that the SPDX exporter rejects an sbom, with a given message.

An action failure can't be observed by an analysis test, so the rule runs the
exporter itself and fails its own action unless the exporter fails with the
expected message. Building the target is the test.
"""

load("@supply_chain_tools//sbom:providers.bzl", "SbomInfo")

def _spdx_expected_failure_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".ok")
    ctx.actions.run_shell(
        outputs = [out],
        inputs = depset(
            transitive = [
                ctx.attr._spdx[DefaultInfo].data_runfiles.files,
                ctx.attr.sbom[DefaultInfo].files,
            ],
        ),
        tools = [ctx.attr._spdx[DefaultInfo].files_to_run],
        command = """
if "$1" --graph "$2" --classifications "$3" --out "$4.report" --format "$5" 2> "$4.stderr"; then
  echo "the exporter was expected to fail on $2" >&2
  exit 1
fi
if ! grep -qF -- "$6" "$4.stderr"; then
  echo "the exporter failed without the message: $6" >&2
  cat "$4.stderr" >&2
  exit 1
fi
: > "$4"
""",
        arguments = [
            ctx.executable._spdx.path,
            ctx.attr.sbom[SbomInfo].graph.path,
            ctx.attr.sbom[SbomInfo].classifications.path,
            out.path,
            ctx.attr.format,
            ctx.attr.message,
        ],
    )
    return [DefaultInfo(files = depset([out]))]

spdx_expected_failure = rule(
    _spdx_expected_failure_impl,
    attrs = {
        "sbom": attr.label(mandatory = True, providers = [SbomInfo]),
        "format": attr.string(mandatory = True),
        "message": attr.string(mandatory = True, doc = "Text the exporter must print when it fails."),
        "_spdx": attr.label(
            default = "@supply-chain-go//cmd/spdx",
            executable = True,
            cfg = "exec",
        ),
    },
)
