"""Declares rule `license`."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

def _license_impl(ctx):
    metadata = ctx.actions.declare_file("{}.package-metadata.attribute.build.bazel.supply-chain.license.json".format(ctx.attr.name))

    ctx.actions.write(
        output = metadata,
        content = json.encode({
            "kind": "build.bazel.supply-chain.license",
            "label": str(ctx.label),
        }),
    )

    return [
        DefaultInfo(
            files = depset(
                direct = [
                    metadata,
                ],
            ),
        ),
        PackageAttributeInfo(
            metadata = metadata,
            kind = "build.bazel.supply-chain.license",
            files = depset(
                direct = [
                    metadata,
                ],
                transitive = [
                    # TODO(yannic): Define files to propagate here.
                ],
            ),
        ),
    ]

_license = rule(
    implementation = _license_impl,
    attrs = {
        # TODO
    },
    provides = [
        PackageAttributeInfo,
    ],
    doc = """
""".strip(),
)

# TODO(yannic): Remove wrapper when Starlark rules support `deleted_attributes`.
def license(
        # Disallow unnamed attributes.
        *,
        # `_license` attributes.
        name,
        # Common attributes (subset since this target is non-configurable).
        visibility = None):
    _license(
        # `_license` attributes.
        name = name,

        # Common attributes.
        visibility = visibility,
        applicable_licenses = [],
    )
