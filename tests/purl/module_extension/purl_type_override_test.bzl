"""Analysis tests for PURL type override registration."""

load("@package_metadata//purl:providers.bzl", "PurlTypeInfo")
load("@rules_testing//lib:analysis_test.bzl", "analysis_test")

_TestSpecInfo = provider(
    fields = {
        "info": "PurlTypeInfo from the registered override target.",
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

def _purl_type_override_test_impl(env, target):
    info = target[_TestSpecInfo].info

    err = info.validate(
        type = "apk",
        namespace = "alpine",
        name = "libcrypto3",
        version = "3.5.5-r0",
        qualifiers = {
            "arch": "x86_64",
            "distro": "alpine-edge",
            "upstream": "openssl",
        },
        subpath = None,
    )
    env.expect.that_bool(err == None).equals(True)

    err = info.validate(
        type = "apk",
        namespace = "alpine",
        name = "libcrypto3",
        version = "3.5.5-r0",
        qualifiers = {
            "origin": "openssl",
        },
        subpath = None,
    )
    env.expect.that_str(err).equals("APK qualifier 'origin' is not supported")

    purl = info.normalize(
        type = "APK",
        namespace = "Alpine",
        name = "LibCrypto3",
        version = "3.5.5-r0",
        qualifiers = {
            "upstream": "openssl",
        },
        subpath = None,
    )
    env.expect.that_str(purl.type).equals("apk")
    env.expect.that_str(purl.namespace).equals("alpine")
    env.expect.that_str(purl.name).equals("libcrypto3")

def purl_type_override_test(*, name, target):
    _test_spec(
        name = "{}_spec".format(name),
        target = target,
    )

    analysis_test(
        name = name,
        target = ":{}_spec".format(name),
        impl = _purl_type_override_test_impl,
    )
