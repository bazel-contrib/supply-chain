"""Analysis test for root-module PURL type override precedence."""

load("@package_metadata//purl:providers.bzl", "PurlTypeInfo")
load("@rules_testing//lib:analysis_test.bzl", "analysis_test")

_TestSpecInfo = provider(
    fields = {
        "info": "PurlTypeInfo resolved from the generated registry.",
    },
)

def _test_spec_impl(ctx):
    return [
        _TestSpecInfo(
            info = ctx.attr.target[PurlTypeInfo],
        ),
    ]

_test_spec = rule(
    implementation = _test_spec_impl,
    attrs = {
        "target": attr.label(
            mandatory = True,
            providers = [
                PurlTypeInfo,
            ],
        ),
    },
)

def _root_override_precedence_test_impl(env, target):
    info = target[_TestSpecInfo].info

    err = info.validate(
        type = "apko",
        namespace = "Chainguard",
        name = "WolfiImage",
        version = "1.0.0",
        qualifiers = None,
        subpath = None,
    )
    env.expect.that_bool(err == None).equals(True)

    purl = info.normalize(
        type = "APKO",
        namespace = "Chainguard",
        name = "WolfiImage",
        version = "1.0.0",
        qualifiers = None,
        subpath = None,
    )
    env.expect.that_str(purl.type).equals("apko")
    env.expect.that_str(purl.namespace).equals("chainguard")
    env.expect.that_str(purl.name).equals("wolfiimage")

def root_override_precedence_test(*, name, target):
    _test_spec(
        name = "{}_spec".format(name),
        target = target,
    )

    analysis_test(
        name = name,
        target = ":{}_spec".format(name),
        impl = _root_override_precedence_test_impl,
    )
