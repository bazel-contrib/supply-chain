load("@bazel_skylib//rules:diff_test.bzl", "diff_test")
load("@package_metadata//attributes:copyright_notice.bzl", "copyright_notice")
load("@package_metadata//attributes:cpe.bzl", "cpe")
load("@package_metadata//attributes:integrity.bzl", "integrity")
load("@package_metadata//attributes:patch.bzl", "patch")
load("@package_metadata//attributes:patch_set.bzl", "patch_set")

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

[
    [
        integrity(
            name = "integrity.{}".format(name),
            values = values,
        ),
        diff_test(
            name = "integrity.{}.test".format(name),
            file1 = ":integrity.{}".format(name),
            file2 = "testdata/integrity/{}.json".format(name),
        ),
    ]
    for name, values in {
        "valid-1": [
            "sha1-he0IF6+DokrY2mjCtQlN5pgzmDw=",
            "sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q=",
        ],
    }.items()
]

patch(
    name = "patch.valid-1",
    description = "Fixes a buffer overflow in the parser.",
    integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
    resolves = ["CVE-2024-12345"],
    type = "backport",
    url = "https://example.com/patches/fix_overflow.patch",
)

diff_test(
    name = "patch.valid-1.test",
    file1 = ":patch.valid-1",
    file2 = "testdata/patch/valid-1.json",
)

patch_set(
    name = "patch_set.valid-1",
    patches = [
        ":patch.valid-1",
        ":patch.valid-2",
    ],
)

patch(
    name = "patch.valid-2",
    description = "Local tweak.",
    file = "testdata/patch/local_tweak.patch",
    type = "unofficial",
    integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
)

diff_test(
    name = "patch_set.valid-1.test",
    file1 = ":patch_set.valid-1",
    file2 = "testdata/patch_set/valid-1.json",
)
