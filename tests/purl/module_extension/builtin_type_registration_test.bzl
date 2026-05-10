"""Analysis test for built-in PURL types exposed through the extension registry."""

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

def _builtin_type_registration_test_impl(env, target):
    info = target[_TestSpecInfo].info

    err = info.validate(
        type = "npm",
        namespace = "Angular",
        name = "Animations",
        version = "12.3.1",
        qualifiers = None,
        subpath = None,
    )
    env.expect.that_bool(err == None).equals(True)

    purl = info.normalize(
        type = "NPM",
        namespace = "Angular",
        name = "Animations",
        version = "12.3.1",
        qualifiers = None,
        subpath = None,
    )
    env.expect.that_str(purl.type).equals("npm")
    env.expect.that_str(purl.namespace[0]).equals("angular")
    env.expect.that_str(purl.name).equals("animations")

def builtin_type_registration_test(*, name, target):
    _test_spec(
        name = "{}_spec".format(name),
        target = target,
    )

    analysis_test(
        name = name,
        target = ":{}_spec".format(name),
        impl = _builtin_type_registration_test_impl,
    )
