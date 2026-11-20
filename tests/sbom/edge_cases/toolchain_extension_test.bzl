"""Tests for provider declarations and root-module precedence."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load(
    "@supply_chain_tools//gather_metadata:extensions.bzl",
    "merge_toolchain_involvement_declarations",
)

def _declaration(module_name, involvement, is_root = False):
    return struct(
        key = "js_binary|@fixture//:node_type",
        rule_kind = "js_binary",
        toolchain_type = "@fixture//:node_type",
        involvement = involvement,
        module_name = module_name,
        is_root = is_root,
        source = module_name,
    )

def _toolchain_extension_test_impl(ctx):
    env = analysistest.begin(ctx)

    root_wins = merge_toolchain_involvement_declarations([
        _declaration("provider", "build"),
        _declaration("root", "runtime", is_root = True),
    ])
    asserts.equals(env, [], root_wins.conflicts)
    asserts.equals(env, 1, len(root_wins.declarations))
    asserts.equals(env, "runtime", root_wins.declarations[0].involvement)

    non_root_conflict = merge_toolchain_involvement_declarations([
        _declaration("provider_a", "build"),
        _declaration("provider_b", "embedded"),
    ])
    asserts.equals(env, 1, len(non_root_conflict.conflicts))
    asserts.true(env, "provider_a" in non_root_conflict.conflicts[0])
    asserts.true(env, "provider_b" in non_root_conflict.conflicts[0])

    return analysistest.end(env)

toolchain_extension_test = analysistest.make(_toolchain_extension_test_impl)
