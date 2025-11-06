load("@package_metadata//:defs.bzl", "package_metadata")
load("@package_metadata_extensions//rules:package_metadata_override.bzl", "package_metadata_override")
load("//metadata-extensions/rules/package_metadata_override:target_group_test.bzl", "target_group_test")

package_metadata(
    name = "metadata",
    purl = "pkg:bazel/foo",
)

package_metadata_override(
    name = "same_repo",
    metadata = ":metadata",
    targets = [
        "//a/b/...",
        "//b/...",
        "-//b/a/...",
        "//c/d",
        "//c:e",
    ],
)

target_group_test(
    name = "same_repo_test",
    excluded = [
        "//a/c",
        "//a/c/d/c",
        "//b/a:x",
        "//b/a/c/d:y",
        "//b/a:a",
        "//b/a/c:foo",
    ],
    included = [
        "//a/b:c",
        "//a/b/c",
        "//a/b/c:c",
        "//a/b/c/d/e/f",
        "//c:e",
        "//c/d",
        "//c/d:x",
        "//c/d:y",
        "//b/c/d:y",
    ],
    target = ":same_repo",
)

package_metadata_override(
    name = "other_repo",
    metadata = ":metadata",
    targets = [
        "@package_metadata//a/b/...",
        "@package_metadata//b/...",
        "-@package_metadata//b/a/...",
        "@package_metadata//c/d",
        "@package_metadata//c:e",
    ],
)

target_group_test(
    name = "other_repo_test",
    excluded = [
        "@package_metadata//a/c",
        "@package_metadata//a/c/d/c",
        "@package_metadata//b/a:x",
        "@package_metadata//b/a/c/d:y",
        "@package_metadata//b/a:a",
        "@package_metadata//b/a/c:foo",
    ],
    included = [
        "@package_metadata//a/b:c",
        "@package_metadata//a/b/c",
        "@package_metadata//a/b/c:c",
        "@package_metadata//a/b/c/d/e/f",
        "@package_metadata//c:e",
        "@package_metadata//c/d",
        "@package_metadata//c/d:x",
        "@package_metadata//c/d:y",
        "@package_metadata//b/c/d:y",
    ],
    target = ":other_repo",
)

package_metadata_override(
    name = "mixed_repo",
    metadata = ":metadata",
    targets = [
        "@package_metadata//a/b/...",
        "@package_metadata//b/...",
        "-@package_metadata//b/a/...",
        "@package_metadata//c/d",
        "@package_metadata//c:e",
        "//a/b/...",
        "//b/...",
        "-//b/a/...",
        "//c/d",
        "//c:e",
        "@package_metadata_extensions//a/b/...",
        "@package_metadata_extensions//b/...",
        "-@package_metadata_extensions//b/a/...",
        "@package_metadata_extensions//c/d",
        "@package_metadata_extensions//c:e",
    ],
)

target_group_test(
    name = "mixed_repo_test",
    excluded = [
        "@package_metadata//a/c",
        "@package_metadata//a/c/d/c",
        "@package_metadata//b/a:x",
        "@package_metadata//b/a/c/d:y",
        "@package_metadata//b/a:a",
        "@package_metadata//b/a/c:foo",
        "//a/c",
        "//a/c/d/c",
        "//b/a:x",
        "//b/a/c/d:y",
        "//b/a:a",
        "//b/a/c:foo",
        "@package_metadata_extensions//a/c",
        "@package_metadata_extensions//a/c/d/c",
        "@package_metadata_extensions//b/a:x",
        "@package_metadata_extensions//b/a/c/d:y",
        "@package_metadata_extensions//b/a:a",
        "@package_metadata_extensions//b/a/c:foo",
    ],
    included = [
        "@package_metadata//a/b:c",
        "@package_metadata//a/b/c",
        "@package_metadata//a/b/c:c",
        "@package_metadata//a/b/c/d/e/f",
        "@package_metadata//c:e",
        "@package_metadata//c/d",
        "@package_metadata//c/d:x",
        "@package_metadata//c/d:y",
        "@package_metadata//b/c/d:y",
        "//a/b:c",
        "//a/b/c",
        "//a/b/c:c",
        "//a/b/c/d/e/f",
        "//c:e",
        "//c/d",
        "//c/d:x",
        "//c/d:y",
        "//b/c/d:y",
        "@package_metadata_extensions//a/b:c",
        "@package_metadata_extensions//a/b/c",
        "@package_metadata_extensions//a/b/c:c",
        "@package_metadata_extensions//a/b/c/d/e/f",
        "@package_metadata_extensions//c:e",
        "@package_metadata_extensions//c/d",
        "@package_metadata_extensions//c/d:x",
        "@package_metadata_extensions//c/d:y",
        "@package_metadata_extensions//b/c/d:y",
    ],
    target = ":mixed_repo",
)
