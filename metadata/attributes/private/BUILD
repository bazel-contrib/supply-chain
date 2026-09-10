load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("//attributes:copyright_notice.bzl", "copyright_notice")

copyright_notice(
    name = "copyright_notice",
    text = [
        "Copyright © 2026 The authors of this stuff.",
        "Copyright © 1999 A. Cöntributor",
    ],
    tags = ["manual"],
)

diff_test(
    name = "copyright_notice_test",
    file1 = ":copyright_notice",
    file2 = "expected_copyright_notice.json",
)
