load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@stardoc//stardoc:stardoc.bzl", "stardoc")

[
    [
        stardoc(
            name = file,
            out = "{}.generated.md".format(file),
            input = "@package_metadata_extensions//rules:{}.bzl".format(file),
            deps = [
                "@package_metadata_extensions//rules:srcs",
            ],
        ),
        diff_test(
            name = "{}_test".format(file),
            file1 = ":{}".format(file),
            file2 = "{}.md".format(file),
        ),
    ]
    for file in [
        "package_metadata_override",
    ]
]
