"""Tests of `PackageAttributeInfo.allow_multiple` in `package_metadata`."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load("@package_metadata//:defs.bzl", "PackageAttributeInfo")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

def _fake_attribute_impl(ctx):
    output = ctx.actions.declare_file(ctx.label.name + ".json")
    ctx.actions.write(output, "{}")
    return [
        DefaultInfo(files = depset([output])),
        PackageAttributeInfo(
            kind = ctx.attr.kind,
            attributes = output,
            allow_multiple = ctx.attr.allow_multiple,
        ),
    ]

_fake_attribute = rule(
    implementation = _fake_attribute_impl,
    attrs = {
        "allow_multiple": attr.bool(),
        "kind": attr.string(mandatory = True),
    },
)

def _metadata_json_test_impl(ctx):
    env = analysistest.begin(ctx)
    content = json.decode(analysistest.target_actions(env)[0].content)
    attributes = content["attributes"]

    multiple = attributes["test.multiple"]
    asserts.true(env, type(multiple) == "list", "expected a list, got %r" % multiple)
    asserts.equals(env, 2, len(multiple))
    asserts.true(env, multiple[0].endswith("_multiple_a.json"))
    asserts.true(env, multiple[1].endswith("_multiple_b.json"))

    single = attributes["test.single"]
    asserts.true(env, type(single) == "string", "expected a string, got %r" % single)
    return analysistest.end(env)

_metadata_json_test = analysistest.make(_metadata_json_test_impl)

def _duplicate_single_kind_test_impl(ctx):
    env = analysistest.begin(ctx)
    asserts.expect_failure(env, "can only be declared once")
    return analysistest.end(env)

_duplicate_single_kind_test = analysistest.make(
    _duplicate_single_kind_test_impl,
    expect_failure = True,
)

def multiple_attributes_tests(name):
    """Declares the tests of attributes that can be declared multiple times.

    Args:
      name: prefix of the declared targets.
    """
    _fake_attribute(name = name + "_multiple_a", kind = "test.multiple", allow_multiple = True, tags = ["manual"])
    _fake_attribute(name = name + "_multiple_b", kind = "test.multiple", allow_multiple = True, tags = ["manual"])
    _fake_attribute(name = name + "_single", kind = "test.single", tags = ["manual"])
    _fake_attribute(name = name + "_single_again", kind = "test.single", tags = ["manual"])

    package_metadata(
        name = name + "_multiple_metadata",
        attributes = [
            ":" + name + "_multiple_a",
            ":" + name + "_single",
            ":" + name + "_multiple_b",
        ],
        purl = "pkg:generic/multiple",
        tags = ["manual"],
    )
    package_metadata(
        name = name + "_duplicate_metadata",
        attributes = [
            ":" + name + "_single",
            ":" + name + "_single_again",
        ],
        purl = "pkg:generic/duplicate",
        tags = ["manual"],
    )

    _metadata_json_test(
        name = name + "_json_test",
        target_under_test = ":" + name + "_multiple_metadata",
    )
    _duplicate_single_kind_test(
        name = name + "_duplicate_test",
        target_under_test = ":" + name + "_duplicate_metadata",
    )
