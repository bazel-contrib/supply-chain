"""Compares a generated SBOM with a golden file, both rendered by jq.

jq sorts the keys and pretty-prints, so goldens are readable and their diffs
are meaningful. The creation time of an SPDX document is pinned, since it is
the only part of a report that changes between runs.

To update a golden after an intended change:

    bazel build //sbom:<name>_actual && \
        cp bazel-bin/sbom/<name>_actual.json sbom/testdata/<report>.json
"""

load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@jq.bzl//jq:jq.bzl", "jq")

_PIN_CREATION_TIME = 'walk(if type == "object" and has("created") then .created = "1970-01-01T00:00:00Z" else . end)'

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
