"""Checks that an exporter rejects an sbom or one of its inputs, with a given message.

An action failure can't be observed by an analysis test, so the rule runs the
exporter itself and fails its own action unless the exporter fails with the
expected message. Building the target is the test.
"""

load("@supply_chain_tools//sbom:providers.bzl", "SbomInfo")

def _expected_failure_impl(ctx, tool, extra_flag, extra_file):
    out = ctx.actions.declare_file(ctx.label.name + ".ok")
    extra_args = []
    extra_inputs = []
    if extra_file:
        extra_args = [extra_flag, extra_file.path]
        extra_inputs = [extra_file]
    ctx.actions.run_shell(
        outputs = [out],
        inputs = depset(
            extra_inputs,
            transitive = [
                tool[DefaultInfo].data_runfiles.files,
                ctx.attr.sbom[DefaultInfo].files,
            ],
        ),
        tools = [tool[DefaultInfo].files_to_run],
        command = """
tool="$1"; graph="$2"; classifications="$3"; out="$4"; format="$5"; message="$6"; shift 6
if "$tool" --graph "$graph" --classifications "$classifications" --out "$out.report" --format "$format" "$@" 2> "$out.stderr"; then
  echo "the exporter was expected to fail on $graph" >&2
  exit 1
fi
if ! grep -qF -- "$message" "$out.stderr"; then
  echo "the exporter failed without the message: $message" >&2
  cat "$out.stderr" >&2
  exit 1
fi
: > "$out"
""",
        arguments = [
            tool[DefaultInfo].files_to_run.executable.path,
            ctx.attr.sbom[SbomInfo].graph.path,
            ctx.attr.sbom[SbomInfo].classifications.path,
            out.path,
            ctx.attr.format,
            ctx.attr.message,
        ] + extra_args,
    )
    return [DefaultInfo(files = depset([out]))]

def _spdx_expected_failure_impl(ctx):
    return _expected_failure_impl(ctx, ctx.attr._spdx, "--document_namespace_file", ctx.file.document_namespace)

def _cyclonedx_expected_failure_impl(ctx):
    return _expected_failure_impl(ctx, ctx.attr._cyclonedx, "--serial_number_file", ctx.file.serial_number)

_COMMON_ATTRS = {
    "sbom": attr.label(mandatory = True, providers = [SbomInfo]),
    "format": attr.string(mandatory = True),
    "message": attr.string(mandatory = True, doc = "Text the exporter must print when it fails."),
}

spdx_expected_failure = rule(
    _spdx_expected_failure_impl,
    attrs = dict(
        _COMMON_ATTRS,
        document_namespace = attr.label(allow_single_file = True, doc = "File to pass as the document namespace."),
        _spdx = attr.label(
            default = "@supply-chain-go//cmd/spdx",
            executable = True,
            cfg = "exec",
        ),
    ),
)

cyclonedx_expected_failure = rule(
    _cyclonedx_expected_failure_impl,
    attrs = dict(
        _COMMON_ATTRS,
        serial_number = attr.label(allow_single_file = True, doc = "File to pass as the serial number."),
        _cyclonedx = attr.label(
            default = "@supply-chain-go//cmd/cyclonedx",
            executable = True,
            cfg = "exec",
        ),
    ),
)
