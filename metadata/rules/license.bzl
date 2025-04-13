"""Declares rule `license`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")
load("//attributes:license.bzl", "KIND")

visibility("public")

def _license_impl(ctx):
    attributes = {
        "text_path": ctx.file.text.path
    }
    if len(ctx.attr.full_name) > 0:
        attributes["full_name"] = ctx.attr.full_name
    if len(ctx.attr.spdx_short_identifier) > 0:
        attributes["spdx_short_identifier"] = ctx.attr.spdx_short_identifier
    if len(ctx.attr.urls) > 0:
        attributes["urls"] = ctx.attr.urls

    attributes_file = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
    ctx.actions.write(
        attributes_file,
        json.encode(attributes)
    )

    license_content_depset = depset(direct=[ctx.file.text])
    return [
        DefaultInfo(
            files = depset(
                direct = [attributes_file],
                transitive = [license_content_depset],
            ),
        ),
        PackageAttributeInfo(
            kind=KIND, 
            attributes=attributes_file, 
            files=[license_content_depset],
        ),
    ]

license = rule(
    implementation = _license_impl,
    attrs = {
        "full_name": attr.string(
            doc = """
The human-readable name of this license.
""".strip(),
        ),
        "spdx_short_identifier": attr.string(
            doc = """
The [SPDX short identifier](https://spdx.org/licenses/) of this license. 
""".strip(),
        ),
        "urls": attr.string_list(
            doc = """
The URLs pointing at the definition of this license.
""".strip(),
        ),
        "text": attr.label(
            allow_single_file = True,
            doc = """
The [File](https://bazel.build/rules/lib/builtins/File) containting the text of this license.
""".strip(),
        ),
    },
    provides = [
        PackageAttributeInfo,
    ],
    doc = """
Rule for declaring a license.
""".strip(),
)