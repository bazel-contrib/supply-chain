load("//purl/private:tables.bzl", "tables")

visibility([
    "//purl/...",
])

percent_encoding = struct(
    encode = {int(k): v for k, v in tables["percent_encoding"]["encode"].items()},
    tests = tables["percent_encoding"]["tests"],
)
