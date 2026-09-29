load("//purl/private/strings:bytes_test.bzl", "bytes_test")

exports_files(
    [
        "ascii.bzl",
        "bytes.bzl",
        "bytes_test.bzl",
        "chrord.bzl",
        "strings.bzl",
    ],
    visibility = [
        "//purl/private:__pkg__",
    ],
)

filegroup(
    name = "srcs",
    srcs = [
        "ascii.bzl",
        "bytes.bzl",
        "bytes_test.bzl",
        "chrord.bzl",
        "strings.bzl",
    ],
    visibility = [
        "//purl/private:__pkg__",
    ],
)

bytes_test(
    name = "bytes_test",
    from_string_cases = {
        "": "",
        "\t": "9",
        "\n": "10",
        " ": "32",
        # Special characters
        "!@#$%": "33,64,35,36,37",
        "123": "49,50,51",
        "A": "65",
        # é (U+00E9) -> UTF-8: C3 A9 -> ISO-8859-1: 0xC3, 0xA9 (195, 169)
        "café": "99,97,102,195,169",
        # ASCII characters
        "hello": "104,101,108,108,111",
        # UTF-8 multi-byte characters (stored as multiple ISO-8859-1 bytes in Bazel)
        # ü (U+00FC) -> UTF-8: C3 BC -> ISO-8859-1: 0xC3, 0xBC (195, 188)
        "ü": "195,188",
    },
    roundtrip_cases = [
        # ASCII
        "hello world",
        "The quick brown fox",
        "123456789",
        "!@#$%^&*()",
        "",
        " ",
        "\n\t\r",
        # Mixed ASCII and UTF-8
        "Hello, Wörld!",
        "Köln",
        "München",
        "café",
        "naïve",
        # Special characters
        "path/to/file",
        "user@example.com",
        "50%",
    ],
    to_string_cases = {
        "": "",
        # Control characters (now supported via octal escapes)
        "\000": "0",
        "\001": "1",
        "\t": "9",
        "\n": "10",
        "\r": "13",
        " ": "32",
        "A": "65",
        "abc": "97,98,99",
        # Basic ASCII
        "hello": "104,101,108,108,111",
        "\200": "128",
        # High bytes (ISO-8859-1)
        "\377": "255",
    },
)
