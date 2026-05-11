"""Utils to normalize [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/types:requirements.bzl", "TYPE_REQUIREMENTS")
load("@package_metadata_purl_types//:normalizers.bzl", "TYPE_NORMALIZERS")

visibility([
    "//purl/private",
    "//purl/types",
])

def _split_path(value):
    return [segment for segment in value.split("/") if segment] if value else None

def _normalize_qualifiers(qualifiers):
    if not qualifiers:
        return None

    normalized = {}
    for key, value in qualifiers.items():
        if value == None or value == "":
            continue
        normalized[key.lower()] = value

    return normalized if normalized else None

def _namespace_to_segments(namespace):
    if type(namespace) == type([]):
        return namespace
    return _split_path(namespace)

def _subpath_to_segments(subpath):
    if type(subpath) == type([]):
        return subpath
    if not subpath:
        return None

    segments = []
    for segment in subpath.split("/"):
        if segment == "" or segment == "." or segment == "..":
            continue
        segments.append(segment)
    return segments if segments else None

def _normalize_type_specific(*, type, namespace, name, version, qualifiers, subpath):
    normalizer = TYPE_NORMALIZERS.get(type)
    if not normalizer:
        return struct(
            type = type,
            namespace = namespace,
            name = name,
            version = version,
            qualifiers = qualifiers,
            subpath = subpath,
        )

    return normalizer(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )

def _lower_segments(segments):
    return [segment.lower() for segment in segments] if segments else segments

def _normalize_by_definition(type, purl):
    definition = TYPE_REQUIREMENTS.get(type, {})
    return struct(
        type = purl.type,
        namespace = _lower_segments(purl.namespace) if definition.get("lower_namespace") else purl.namespace,
        name = purl.name.lower() if definition.get("lower_name") else purl.name,
        version = purl.version.lower() if definition.get("lower_version") and purl.version else purl.version,
        qualifiers = purl.qualifiers,
        subpath = _lower_segments(purl.subpath) if definition.get("lower_subpath") else purl.subpath,
    )

def normalize(
        *,
        type = None,
        namespace = None,
        name = None,
        version = None,
        qualifiers = {},
        subpath = None):
    if not type:
        return None, "Mandatory property 'type' not set"
    if not name:
        return None, "Mandatory property 'name' not set"

    normalized_type = type.lower()
    purl = _normalize_type_specific(
        type = normalized_type,
        namespace = _namespace_to_segments(namespace),
        name = name,
        version = version,
        qualifiers = _normalize_qualifiers(qualifiers),
        subpath = _subpath_to_segments(subpath),
    )
    return _normalize_by_definition(normalized_type, purl), None
