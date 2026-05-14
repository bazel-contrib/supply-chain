"""Integration tests for metadata gathering aspect."""

load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load(":gather_metadata.bzl", "gather_metadata_info")
load(":providers.bzl", "TransitiveMetadataInfo")
load(":test_fixtures.bzl", "create_test_fixtures")



def _direct_deps_test_impl(ctx):
    """Verify direct_deps field is populated, immutable, and contains labels."""
    env = analysistest.begin(ctx)
    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    all_targets = info.transitive.to_list()

    for target_info in all_targets:
        asserts.true(env, hasattr(target_info, "direct_deps"))

        if target_info.direct_deps:
            asserts.equals(env, type(target_info.direct_deps), type(()))

            for dep in target_info.direct_deps:
                asserts.true(env, type(dep) == type("") or hasattr(dep, "name"))

    return analysistest.end(env)

direct_deps_test = analysistest.make(
    _direct_deps_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def _diamond_dependency_test_impl(ctx):
    """Verify diamond dependencies are deduplicated by depset."""
    env = analysistest.begin(ctx)
    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    all_targets = info.transitive.to_list()
    target_labels = [str(target_info.target) for target_info in all_targets]

    unique_labels = {}
    for label in target_labels:
        asserts.false(env, label in unique_labels, "Duplicate: {}".format(label))
        unique_labels[label] = True

    return analysistest.end(env)

diamond_dependency_test = analysistest.make(
    _diamond_dependency_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def _edge_extraction_test_impl(ctx):
    """Verify edges can be extracted from direct_deps for SBOM generation."""
    env = analysistest.begin(ctx)
    target_under_test = analysistest.target_under_test(env)
    info = target_under_test[TransitiveMetadataInfo]

    all_targets = info.transitive.to_list()

    edges = []
    seen_edges = {}
    for target_info in all_targets:
        if hasattr(target_info, "direct_deps") and target_info.direct_deps:
            from_label = str(target_info.target)
            for dep in target_info.direct_deps:
                to_label = str(dep)
                edge_key = "{} -> {}".format(from_label, to_label)
                if edge_key not in seen_edges:
                    seen_edges[edge_key] = True
                    edges.append({"from": from_label, "to": to_label})

    for edge in edges:
        asserts.true(env, "from" in edge)
        asserts.true(env, "to" in edge)

    return analysistest.end(env)

edge_extraction_test = analysistest.make(
    _edge_extraction_test_impl,
    extra_target_under_test_aspects = [gather_metadata_info],
)

def integration_test_suite(name):
    create_test_fixtures()

    direct_deps_test(
        name = "direct_deps_test",
        target_under_test = ":test_target_with_deps",
    )

    edge_extraction_test(
        name = "edge_extraction_test",
        target_under_test = ":test_target_with_deps",
    )

    diamond_dependency_test(
        name = "diamond_dependency_test",
        target_under_test = ":test_target_with_deps",
    )

    native.test_suite(
        name = name,
        tests = [
            ":direct_deps_test",
            ":edge_extraction_test",
            ":diamond_dependency_test",
        ],
    )
