"""Tests for provider declarations and root-module precedence."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load(
    "@supply_chain_tools//gather_metadata:extensions.bzl",
    "merge_relationship_declarations",
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

def _relationship_declaration(
        module_name,
        relationship,
        is_root = False,
        rule_module = "",
        rule_kind = "pkg_tar",
        attribute = "srcs",
        requires_attributes = []):
    requires_attributes = sorted(requires_attributes)
    return struct(
        key = "{}|{}|{}|{}".format(rule_module, rule_kind, attribute, ",".join(requires_attributes)),
        rule_module = rule_module,
        rule_kind = rule_kind,
        attribute = attribute,
        requires_attributes = requires_attributes,
        relationship = relationship,
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

    root_wins = merge_relationship_declarations([
        _relationship_declaration("provider", "depends_on"),
        _relationship_declaration("root", "contains", is_root = True),
    ])
    asserts.equals(env, [], root_wins.conflicts)
    asserts.equals(env, 1, len(root_wins.declarations))
    asserts.equals(env, "contains", root_wins.declarations[0].relationship)

    identical = merge_relationship_declarations([
        _relationship_declaration("provider_a", "contains"),
        _relationship_declaration("provider_b", "contains"),
    ])
    asserts.equals(env, [], identical.conflicts)
    asserts.equals(env, 1, len(identical.declarations))

    relationship_conflict = merge_relationship_declarations([
        _relationship_declaration("provider_a", "contains"),
        _relationship_declaration("provider_b", "static_link"),
    ])
    asserts.equals(env, 1, len(relationship_conflict.conflicts))
    asserts.true(env, "provider_a" in relationship_conflict.conflicts[0])
    asserts.true(env, "provider_b" in relationship_conflict.conflicts[0])

    separated_relationships = merge_relationship_declarations([
        _relationship_declaration(
            "rules_img",
            "contains",
            rule_module = "rules_img",
            rule_kind = "image_layer",
            attribute = "srcs",
        ),
        _relationship_declaration(
            "rules_pkg",
            "contains",
            rule_module = "rules_pkg",
            rule_kind = "image_layer",
            attribute = "srcs",
        ),
    ])
    asserts.equals(env, [], separated_relationships.conflicts)
    asserts.equals(env, 2, len(separated_relationships.declarations))

    return analysistest.end(env)

toolchain_extension_test = analysistest.make(_toolchain_extension_test_impl)
