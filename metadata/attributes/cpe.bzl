"""Declares rule `cpe`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

visibility("public")

# The `kind` of the `PackageAttributeInfo` provided by `cpe` targets.
KIND = "build.bazel.attribute.cpe"

def _cpe_impl(ctx):
    output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        output = output,
        content = json.encode(ctx.attr.identifiers),
    )
    return [
        DefaultInfo(
            files = depset(direct = [output]),
        ),
        PackageAttributeInfo(
            kind = KIND,
            attributes = output,
        ),
    ]

# Documentation lives on the public `cpe` macro below: Stardoc renders the
# macro, not this private rule.
_cpe = rule(
    implementation = _cpe_impl,
    attrs = {
        "identifiers": attr.string_list(
            mandatory = True,
        ),
    },
)

def cpe(
        # Disallow unnamed attributes.
        *,
        # `_cpe` attributes.
        name,
        identifiers,
        # Common attributes (subset since this target is non-configurable).
        tags = None,
        visibility = None):
    """Rule for declaring the CPEs of a package or target.

    This is typically passed to the `attributes` of a `package_metadata` target.

    `identifiers` entries are assumed to be UTF-8 encoded. Tools for building SBOMs
    presume that.

    Each identifier is a CPE 2.2 URI (`cpe:/a:acme:widget:1.0`) or a CPE 2.3
    formatted string (`cpe:2.3:a:acme:widget:1.0:*:*:*:*:*:*:*`). SBOM generators
    reject anything else. They report every identifier: in SPDX 2.3 as an
    external reference of the `SECURITY` category, of the type `cpe22Type` or
    `cpe23Type`, and in SPDX 3.0.1 as an external identifier of the type `cpe22` or
    `cpe23`. A CycloneDX component has a single `cpe`, so it gets the first
    identifier, and the other ones are properties named `bazel:cpe`.

    Note: a tool that only reads the `cpe` of a CycloneDX component, such as
    Dependency-Track, never sees the `bazel:cpe` properties, so it matches a
    component against a single CPE. Vulnerability reports can then under-report.
    The NVD may list the CVEs of one product under several CPEs, for example
    under different vendor names (`cpe:2.3:a:teluu:pjsip` and
    `cpe:2.3:a:pjsip:pjsip` for PJSIP), and a component that carries only one of
    them misses the CVEs listed under the others. Listing the same component
    several times in the report, each time with one of the CPEs, lets such a
    tool match all of them. The cost is that the component appears several
    times, and that a CVE listed under more than one of those CPEs can be
    reported more than once.

    Usage:

    ```starlark
    load("@package_metadata//attributes:cpe.bzl", "cpe")
    load("@package_metadata//purl:purl.bzl", "purl")
    load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

    package_metadata(
        name = "package_metadata",
        attributes = [
            ":cpe",
        ],
        purl = purl.builder().type("generic").name("ntp").version("4.2.8").build(),
        visibility = ["//visibility:public"],
    )

    cpe(
        name = "cpe",
        identifiers = [
            "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*",
        ],
    )
    ```

    Format: the `PackageAttributeInfo` of this rule has `kind`
    `build.bazel.attribute.cpe`, and its `attributes` file holds the `identifiers`
    entries as a JSON array of strings, encoded as UTF-8 and written without
    insignificant whitespace:

    ```json
    [
        "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*"
    ]
    ```

    Args:
      name: A unique name for this target.
      identifiers: Required. One or more
        [CPE](https://en.wikipedia.org/wiki/Common_Platform_Enumeration)
        identifiers of the package. UTF-8 encoding.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """

    _cpe(
        # `_cpe` attributes.
        name = name,
        identifiers = identifiers,

        # Common attributes.
        tags = tags,
        visibility = visibility,
        # This should be package_metadata, but we use the legacy name
        # to support bazel 7.
        applicable_licenses = [],
    )
