"""Compares a generated SBOM with a golden file, both rendered by jq.

jq sorts the keys and pretty-prints, so goldens are readable and their diffs
are meaningful. The creation time of an SPDX document and the timestamp of a
CycloneDX one are pinned, since they are the only parts of a report that change
between runs.

To update a golden after an intended change:

    bazel build //sbom:<name>_actual && \
        cp bazel-bin/sbom/<name>_actual.json sbom/testdata/<report>.json
"""

load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@bazel_skylib//rules:write_file.bzl", "write_file")
load("@jq.bzl//jq:jq.bzl", "jq")

_PIN_CREATION_TIME = 'walk(if type == "object" and has("created") then .created = "1970-01-01T00:00:00Z" elif type == "object" and has("bomFormat") then .metadata.timestamp = "1970-01-01T00:00:00Z" else . end)'

def sbom_golden_test(name, report, golden, **kwargs):
    """Tests that a `spdx` or `cyclonedx` report matches a golden file.

    Args:
      name: test name; the rendered report is `<name>_actual`.
      report: the `spdx` or `cyclonedx` target.
      golden: the golden file, already rendered by jq.
      **kwargs: forwarded to the test.
    """
    jq(
        name = name + "_actual",
        srcs = [report],
        args = ["-S"],
        filter = _PIN_CREATION_TIME,
        out = name + "_actual.json",
    )
    diff_test(
        name = name,
        file1 = golden,
        file2 = ":" + name + "_actual",
        failure_message = "The report differs from its golden file; see golden_test.bzl to update it.",
        **kwargs
    )

_EPOCH = "1970-01-01T00:00:00Z"

def sbom_timestamp_test(name, report, path, **kwargs):
    """Tests that the time of a report is the Unix epoch exactly when the build is not stamped.

    The golden tests pin the time of a report, so they do not tell whether it is
    deterministic. This one looks at the report as it is generated, and passes
    with and without `--stamp`.

    Args:
      name: test name.
      report: the `spdx` or `cyclonedx` target.
      path: jq expression of the time in the report.
      **kwargs: forwarded to the test.
    """
    jq(
        name = name + "_actual",
        srcs = [report],
        args = ["-r"],
        filter = select({
            "@supply_chain_tools//sbom:stamp": "(%s) != \"%s\"" % (path, _EPOCH),
            "//conditions:default": "(%s) == \"%s\"" % (path, _EPOCH),
        }),
        out = name + "_actual.txt",
    )
    write_file(
        name = name + "_expected",
        out = name + "_expected.txt",
        content = ["true", ""],
    )
    diff_test(
        name = name,
        file1 = ":" + name + "_expected",
        file2 = ":" + name + "_actual",
        failure_message = "The time of the report is not the Unix epoch exactly when the build is not stamped.",
        **kwargs
    )
