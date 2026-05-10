"""Analysis test for PURL type registrations contributed by submodules."""

load("@package_metadata//purl:providers.bzl", "PurlTypeInfo")
load("@rules_testing//lib:analysis_test.bzl", "analysis_test")

_TestSpecInfo = provider(
    fields = {
        "info": "PurlTypeInfo from the registered submodule target.",
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

def _ansible_type_registration_test_impl(env, target):
    info = target[_TestSpecInfo].info

    err = info.validate(
        type = "ansible",
        namespace = "Community",
        name = "General",
        version = "12.5.0",
        qualifiers = {
            "vcs_url": "git+https://github.com/ansible-collections/community.general.git,12.5.0",
        },
        subpath = None,
    )
    env.expect.that_bool(err == None).equals(True)

    err = info.validate(
        type = "ansible",
        namespace = "community",
        name = "general",
        version = "12.5.0",
        qualifiers = {
            "origin": "galaxy",
        },
        subpath = None,
    )
    env.expect.that_str(err).equals("Ansible qualifier 'origin' is not supported")

    purl = info.normalize(
        type = "ANSIBLE",
        namespace = "Community",
        name = "General",
        version = "12.5.0",
        qualifiers = {
            "packaging": "rpm",
        },
        subpath = None,
    )
    env.expect.that_str(purl.type).equals("ansible")
    env.expect.that_str(purl.namespace).equals("community")
    env.expect.that_str(purl.name).equals("general")

def ansible_type_registration_test(*, name, target):
    _test_spec(
        name = "{}_spec".format(name),
        target = target,
    )

    analysis_test(
        name = name,
        target = ":{}_spec".format(name),
        impl = _ansible_type_registration_test_impl,
    )
