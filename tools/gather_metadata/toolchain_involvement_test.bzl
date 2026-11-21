"""Unit tests for toolchain involvement normalization and graph output."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load(":toolchain_involvement.bzl", "normalize_toolchain_type")
load(":providers.bzl", "TargetWithMetadataInfo", "TransitiveMetadataInfo")
load(":serialization.bzl", "metadata_info_to_json")

_DeclaredInfo = provider(fields = ["declarations"])

def _declared_rule_impl(ctx):
    return [
        DefaultInfo(),
        _DeclaredInfo(declarations = ctx.attr._supply_chain_toolchain_involvement),
    ]

_declared_rule = rule(
    implementation = _declared_rule_impl,
    attrs = {
        "_supply_chain_toolchain_involvement": attr.label_keyed_string_dict(default = {
            Label("@bazel_tools//tools/jdk:toolchain_type"): "build",
        }),
    },
)

def _toolchain_involvement_test_impl(ctx):
    env = analysistest.begin(ctx)

    asserts.equals(env, "@rules_java//x:y", normalize_toolchain_type("@@rules_java+//x:y"))
    asserts.equals(env, "@rules_java//x:y", normalize_toolchain_type("@@rules_java~1.0//x:y"))
    asserts.equals(env, "//x:y", normalize_toolchain_type("@@//x:y"))
    asserts.equals(env, "@bazel_tools//tools/jdk:toolchain_type", normalize_toolchain_type("@@bazel_tools//tools/jdk:toolchain_type"))

    declarations = analysistest.target_under_test(env)[_DeclaredInfo].declarations
    declaration_key = declarations.keys()[0]
    asserts.equals(
        env,
        "build",
        declarations[declaration_key],
    )
    asserts.equals(env, Label("@bazel_tools//tools/jdk:toolchain_type"), declaration_key.label)

    root = Label("//:toolchain_involvement_test_root")
    dependency = Label("//:toolchain_involvement_test_dep")
    target_info = TargetWithMetadataInfo(
        target = root,
        metadata = depset(),
        direct_deps = (dependency, dependency),
        direct_dep_involvements = ("runtime", "build"),
    )
    serialized = metadata_info_to_json(TransitiveMetadataInfo(
        transitive = depset([target_info]),
        top_level_target = root,
        is_exec_config = False,
    ))
    edges = json.decode(serialized[0])["edges"]
    asserts.equals(env, 2, len(edges), "Distinct involvement values must retain parallel edges")
    asserts.false(env, "involvement" in edges[0], "Runtime edges omit the optional field")
    asserts.equals(env, "build", edges[1].get("involvement"))

    return analysistest.end(env)

toolchain_involvement_test = analysistest.make(_toolchain_involvement_test_impl)

def toolchain_involvement_test_suite(name):
    toolchain_involvement_test(
        name = "toolchain_involvement_test",
        target_under_test = ":toolchain_involvement_test_subject",
    )
    _declared_rule(name = "toolchain_involvement_test_subject")
    native.test_suite(
        name = name,
        tests = [":toolchain_involvement_test"],
    )
