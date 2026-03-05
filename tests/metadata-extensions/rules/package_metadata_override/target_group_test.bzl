load("@package_metadata//providers:package_metadata_override_info.bzl", "PackageMetadataOverrideInfo")
load("@rules_testing//lib:analysis_test.bzl", "analysis_test")

_TestSpecInfo = provider(
    fields = [
        "info",
        "included",
        "excluded",
    ],
)

def _test_spec_rule_impl(ctx):
    return [
        _TestSpecInfo(
            info = ctx.attr.target[PackageMetadataOverrideInfo],
            included = ctx.attr.included,
            excluded = ctx.attr.excluded,
        ),
    ]

_test_spec_rule = rule(
    implementation = _test_spec_rule_impl,
    attrs = {
        "excluded": attr.string_list(
            mandatory = True,
        ),
        "included": attr.string_list(
            mandatory = True,
        ),
        "target": attr.label(
            mandatory = True,
            providers = [
                PackageMetadataOverrideInfo,
            ],
        ),
    },
    provides = [
        _TestSpecInfo,
    ],
)

def _test_spec_impl(name, target, included, excluded, visibility):
    _test_spec_rule(
        name = name,
        target = target,
        included = [str(label) for label in included],
        excluded = [str(label) for label in excluded],
        visibility = visibility,
    )

_test_spec = macro(
    implementation = _test_spec_impl,
    attrs = {
        "excluded": attr.label_list(
            mandatory = True,
            configurable = False,
        ),
        "included": attr.label_list(
            mandatory = True,
            configurable = False,
        ),
        "target": attr.label(
            mandatory = True,
            configurable = False,
            providers = [
                PackageMetadataOverrideInfo,
            ],
        ),
    },
)

def _target_group_test_impl(env, target):
    spec = target[_TestSpecInfo]

    for label in spec.included:
        env.expect.that_bool(spec.info.packages.contains(Label(label))).equals(True)
    for label in spec.excluded:
        env.expect.that_bool(spec.info.packages.contains(Label(label))).equals(False)

def target_group_test(*, name, target, included, excluded):
    _test_spec(
        name = "{}_spec".format(name),
        target = target,
        included = included,
        excluded = excluded,
    )

    analysis_test(
        name = name,
        target = ":{}_spec".format(name),
        impl = _target_group_test_impl,
    )
