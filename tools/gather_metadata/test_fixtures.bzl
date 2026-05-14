"""Test fixtures for metadata gathering tests."""

def create_test_fixtures():
    """Create test fixture targets for integration tests."""

    native.filegroup(
        name = "test_leaf_dep_d",
        srcs = [],
        tags = ["manual"],
    )

    native.filegroup(
        name = "test_leaf_dep_e",
        srcs = [],
        tags = ["manual"],
    )

    native.filegroup(
        name = "test_dep_b",
        srcs = [":test_leaf_dep_d"],
        tags = ["manual"],
    )

    native.filegroup(
        name = "test_dep_c",
        srcs = [":test_leaf_dep_d", ":test_leaf_dep_e"],
        tags = ["manual"],
    )

    # Diamond pattern:
    #     test_target_with_deps
    #       /              \
    #    test_dep_b    test_dep_c
    #       \              /    \
    #      test_leaf_dep_d   test_leaf_dep_e
    native.filegroup(
        name = "test_target_with_deps",
        srcs = [":test_dep_b", ":test_dep_c"],
        tags = ["manual"],
    )

    native.filegroup(
        name = "test_simple_target",
        srcs = [],
        tags = ["manual"],
    )

    # Target with tools (should be filtered)
    native.genrule(
        name = "test_genrule_with_tools",
        srcs = [":test_simple_target"],
        outs = ["test_out.txt"],
        cmd = "touch $@",
        tools = [":test_dep_b"],
        tags = ["manual"],
    )
