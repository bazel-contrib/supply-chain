"""Unit tests for toolchain-type normalization and registry lookup."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load(":providers.bzl", "TargetWithMetadataInfo", "TransitiveMetadataInfo")
load(":serialization.bzl", "metadata_info_to_json")
load(
    ":toolchain_involvement.bzl",
    "ToolchainInvolvementInfo",
    "normalize_toolchain_type",
    "toolchain_involvement_for",
)

def _toolchain_involvement_test_impl(ctx):
    env = analysistest.begin(ctx)

    asserts.equals(env, "@rules_java//x:y", normalize_toolchain_type("@@rules_java+//x:y"))
    asserts.equals(env, "@rules_java//x:y", normalize_toolchain_type("@@rules_java~1.0//x:y"))
    asserts.equals(env, "//x:y", normalize_toolchain_type("@@//x:y"))
    asserts.equals(env, "@bazel_tools//tools/jdk:toolchain_type", normalize_toolchain_type("@@bazel_tools//tools/jdk:toolchain_type"))

    java_type = "@bazel_tools//tools/jdk:toolchain_type"
    registry = ToolchainInvolvementInfo(entries = {
        "*|" + java_type: "runtime",
        "cc_binary|" + java_type: "embedded",
    })
    asserts.equals(env, "embedded", toolchain_involvement_for(registry, "cc_binary", java_type))
    asserts.equals(env, "runtime", toolchain_involvement_for(registry, "java_binary", java_type))
    asserts.equals(env, None, toolchain_involvement_for(registry, "java_binary", "//other:type"))

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
        target_under_test = ":test_simple_target",
    )
    native.test_suite(
        name = name,
        tests = [":toolchain_involvement_test"],
    )
