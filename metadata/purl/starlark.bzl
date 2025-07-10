"""Module to detect the encoding of Bazel Starlark files."""

load("//purl:string.bzl", "string")

visibility([
    "//purl/...",
])

def _starlark_encoding_test(*, value, table):
    b = "".join([str(i) for i in string.to_bytes(value)])

    encoding = table.get(b, None)
    if not encoding:
        fail("Unknown encoding for {}".format(value))

    return encoding

def _get_encoding(*detected_encodings):
    encodings = {t: True for t in detected_encodings}
    if len(encodings) > 1:
        fail("Detected inconsistent encoding: {}".format(",".join(detected_encodings)))

    for encoding in encodings.keys():
        return encoding

    fail("Illegal state: no encoding")

_encoding = _get_encoding(
    _starlark_encoding_test(
        value = "München",
        table = {
            "7719518811099104101110": "utf8",
            "7725211099104101110": "ISO-8859-1",
        },
    ),
    _starlark_encoding_test(
        value = "Småland",
        table = {
            "8310919516510897110100": "utf8",
            "8310922910897110100": "ISO-8859-1",
        },
    ),
)

starlark = struct(
    encoding = _encoding,
)
