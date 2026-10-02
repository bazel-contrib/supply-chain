"""Single-target license coverage, policy, and SBOM orchestration."""

load(":cyclonedx.bzl", "cyclonedx")
load(":license_coverage.bzl", "license_coverage")
load(":license_policy.bzl", "license_policy")
load(":sbom.bzl", "sbom")
load(":spdx.bzl", "spdx")

def supply_chain(
        name,
        target,
        policy = None,
        resolutions = None,
        sbom_formats = ["spdx-json"],
        require_root_metadata = True,
        visibility = None):
    """Creates one build target for metadata coverage, policy, and SBOMs.

    Args:
      name: Public filegroup containing the generated reports.
      target: Root Bazel target whose transitive metadata is inspected.
      policy: Optional license_policy allowlist target. When omitted, the
        macro still emits a coverage report and SBOMs.
      resolutions: Optional language-specific resolver JSON target. Its
        package expressions are merged into coverage.
      sbom_formats: Any of spdx-json, spdx-yaml, spdx-tag-value,
        cyclonedx-json, and cyclonedx-xml.
      require_root_metadata: Whether the root target must declare metadata.
      visibility: Visibility for the public filegroup.
    """
    supported_formats = {
        "spdx-json": "spdx",
        "spdx-yaml": "spdx",
        "spdx-tag-value": "spdx",
        "cyclonedx-json": "cyclonedx",
        "cyclonedx-xml": "cyclonedx",
    }
    for format in sbom_formats:
        if format not in supported_formats:
            fail("unsupported SBOM format %r" % format)

    sbom_name = name + "_sbom"
    coverage_name = name + "_license_coverage"
    sbom(
        name = sbom_name,
        target = target,
        require_root_metadata = require_root_metadata,
    )
    license_coverage(
        name = coverage_name,
        out = coverage_name + ".json",
        resolutions = resolutions,
        sbom = ":" + sbom_name,
    )

    outputs = [":" + coverage_name]
    if policy != None:
        policy_name = name + "_license_policy"
        license_policy(
            name = policy_name,
            allowlist = policy,
            out = policy_name + ".json",
            report = ":" + coverage_name,
        )
        outputs.append(":" + policy_name)

    for index, format in enumerate(sbom_formats):
        output_name = "%s_%d" % (name, index)
        if supported_formats[format] == "spdx":
            spdx(
                name = output_name,
                coverage = ":" + coverage_name,
                format = {
                    "spdx-json": "json",
                    "spdx-yaml": "yaml",
                    "spdx-tag-value": "tag-value",
                }[format],
                out = "%s.%s" % (name, {
                    "spdx-json": "spdx.json",
                    "spdx-yaml": "spdx.yaml",
                    "spdx-tag-value": "spdx.tag-value",
                }[format]),
                sbom = ":" + sbom_name,
            )
        else:
            cyclonedx(
                name = output_name,
                coverage = ":" + coverage_name,
                format = "xml" if format == "cyclonedx-xml" else "json",
                out = "%s.%s" % (name, "cyclonedx.xml" if format == "cyclonedx-xml" else "cyclonedx.json"),
                sbom = ":" + sbom_name,
            )
        outputs.append(":" + output_name)

    native.filegroup(
        name = name,
        srcs = outputs,
        visibility = visibility,
    )
