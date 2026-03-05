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
    # Spec §5: Validate required fields are present.
    if not type:
        return "Mandatory property 'type' not set"
    if not name:
        return "Mandatory property 'name' not set"

    # TODO(yannic): Implement type-specific validation.

    return None
