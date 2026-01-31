"""Utils to validate [purl](https://github.com/package-url/purl-spec)s."""

visibility([
    "//purl/private",
])

def validate(
        *,
        type = None,
        namespace = None,
        name = None,
        version = None,
        qualifiers = {},
        subpath = None):
    if not type:
        fail("Mandatory property 'type' not set")

    # TODO(yannic): Implement type-specific validation.
