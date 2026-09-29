load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("//attributes:copyright_notice.bzl", "copyright_notice")
load("//attributes:cpe.bzl", "cpe")

[
    [
        copyright_notice(
            name = "copyright_notice.{}".format(name),
            text = text,
        ),
        diff_test(
            name = "copyright_notice.{}.test".format(name),
            file1 = ":copyright_notice.{}".format(name),
            file2 = "testdata/copyright_notice/{}.json".format(name),
        ),
    ]
    for name, text in {
        "valid-1": [
            "Copyright © 2026 The authors of this stuff.",
            "Copyright © 1999 A. Cöntributor",
        ],
        "valid-2": [
            "Copyright © 1999 A. Cöntributor",
        ],
        "valid-3": [
            "Copyright © 2026 The authors of this stuff.",
        ],
    }.items()
]

[
    [
        cpe(
            name = "cpe.{}".format(name),
            identifiers = identifiers,
        ),
        diff_test(
            name = "cpe.{}.test".format(name),
            file1 = ":cpe.{}".format(name),
            file2 = "testdata/cpe/{}.json".format(name),
        ),
    ]
    for name, identifiers in {
        "valid-1": [
            "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*",
        ],
    }.items()
]
