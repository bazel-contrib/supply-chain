"""Utils to normalize [purl](https://github.com/package-url/purl-spec)s."""

visibility([
    "//purl/private",
])

def normalize(
        *,
        type = None,
        namespace = None,
        name = None,
        version = None,
        qualifiers = {},
        subpath = None):
    if not type:
        None, "Mandatory property 'type' not set"

    # TODO(yannic): Implement normalization.

    purl = struct(
        type = type,
        namespace = [segment for segment in namespace.split("/") if segment] if namespace else None,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = [segment for segment in subpath.split("/") if segment] if subpath else None,
    )
    return purl, None
