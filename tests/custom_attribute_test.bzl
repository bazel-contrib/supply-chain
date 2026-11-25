"""End-to-end test of a custom attribute that can be declared multiple times.

Mimics what an organization (or the `patch` attribute) does: declare its own
`PackageAttributeInfo` kind with `allow_multiple = True`, attach it more than
once to a `package_metadata` and check the JSON consumers read.
"""

load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@jq.bzl//jq:jq.bzl", "jq")
load("@package_metadata//:defs.bzl", "PackageAttributeInfo")
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

KIND = "com.example.attribute.patch"

def _patch_impl(ctx):
    output = ctx.actions.declare_file(ctx.label.name + ".patch-attribute.json")
    ctx.actions.write(output, json.encode({"file": ctx.file.file.short_path}))
    return [
        DefaultInfo(files = depset([output])),
        PackageAttributeInfo(
            kind = KIND,
            attributes = output,
            files = [depset([ctx.file.file])],
            allow_multiple = True,
        ),
    ]

_patch = rule(
    implementation = _patch_impl,
    attrs = {
        "file": attr.label(allow_single_file = True, mandatory = True),
    },
    provides = [PackageAttributeInfo],
)

# Keeps only the file names, since the paths contain the configuration.
_BASENAMES = '.attributes |= map_values(if type == "array" then map(split("/") | last) else split("/") | last end)'

def custom_attribute_test(name, golden):
    """Tests the metadata of a package with two custom attributes of one kind.

    Args:
      name: name of the test.
      golden: golden file of the (basename-normalized) metadata JSON.
    """
    native.genrule(
        name = name + "_patch_a_src",
        outs = [name + "_a.patch"],
        cmd = "echo a > $@",
    )
    native.genrule(
        name = name + "_patch_b_src",
        outs = [name + "_b.patch"],
        cmd = "echo b > $@",
    )
    _patch(name = name + "_patch_a", file = ":" + name + "_a.patch")
    _patch(name = name + "_patch_b", file = ":" + name + "_b.patch")
    license(
        name = name + "_license",
        kind = "@package_metadata//licenses/spdx:MIT",
    )
    package_metadata(
        name = name + "_metadata",
        # Declaration order of the patches must be preserved.
        attributes = [
            ":" + name + "_patch_a",
            ":" + name + "_license",
            ":" + name + "_patch_b",
        ],
        purl = "pkg:generic/patched@1.0.0",
    )
    jq(
        name = name + "_actual",
        srcs = [":" + name + "_metadata"],
        args = ["-S"],
        filter = _BASENAMES,
        out = name + "_actual.json",
    )
    diff_test(
        name = name,
        file1 = golden,
        file2 = ":" + name + "_actual",
        failure_message = "The metadata differs from its golden file. Update it with: bazel build //:{0}_actual && cp bazel-bin/{0}_actual.json {1}".format(name, golden),
    )
