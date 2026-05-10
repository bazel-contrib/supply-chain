"""Parser for [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private/normalization:normalization.bzl", "normalize")
load("//purl/private/percent_encoding:percent_encoding.bzl", "percent_decode")
load("//purl/private/validation:validation.bzl", "validate")

visibility([
    "//purl/...",
])

def _split_once_left(value, separator):
    parts = value.split(separator)
    if len(parts) == 1:
        return value, None
    return parts[0], separator.join(parts[1:])

def _split_once_right(value, separator):
    parts = value.split(separator)
    if len(parts) == 1:
        return value, None
    return separator.join(parts[:-1]), parts[-1]

def _strip_leading(value, char):
    for _ in range(len(value)):
        if not value.startswith(char):
            break
        value = value[1:]
    return value

def _strip_trailing(value, char):
    for _ in range(len(value)):
        if not value.endswith(char):
            break
        value = value[:-1]
    return value

def _decode(value, component):
    decoded, err = percent_decode(value)
    if err:
        return None, "{}: {}".format(component, err)
    return decoded, None

def _decode_segments(value, component, discard_dots = False):
    segments = []
    for raw_segment in value.split("/"):
        segment, err = _decode(raw_segment, component)
        if err:
            return None, err
        if segment == "" or (discard_dots and (segment == "." or segment == "..")):
            continue
        if "/" in segment:
            return None, "{} segment contains '/'".format(component)
        segments.append(segment)
    return segments, None

def _parse_qualifiers(value):
    qualifiers = {}
    for pair in value.split("&"):
        key, raw_value = _split_once_left(pair, "=")
        if raw_value == None:
            return None, "Qualifier must be a key=value pair"
        decoded_value, err = _decode(raw_value, "qualifier")
        if err:
            return None, err
        if decoded_value == "":
            continue
        qualifiers[key.lower()] = decoded_value
    return qualifiers if qualifiers else None, None

def _as_dict(purl):
    return {
        "type": purl.type,
        "namespace": "/".join(purl.namespace) if purl.namespace else None,
        "name": purl.name,
        "version": purl.version,
        "qualifiers": purl.qualifiers,
        "subpath": "/".join(purl.subpath) if purl.subpath else None,
    }

def parse(value, strict = True):
    """Parses a PURL string into normalized components."""

    remainder, raw_subpath = _split_once_right(value, "#")
    subpath = None
    if raw_subpath != None:
        subpath_segments, err = _decode_segments(raw_subpath, "subpath", discard_dots = True)
        if err:
            return None, err
        subpath = "/".join(subpath_segments) if subpath_segments else None

    remainder, raw_qualifiers = _split_once_right(remainder, "?")
    qualifiers = None
    if raw_qualifiers != None:
        qualifiers, err = _parse_qualifiers(raw_qualifiers)
        if err:
            return None, err

    scheme, remainder = _split_once_left(remainder, ":")
    if remainder == None:
        return None, "PURL is missing a scheme"
    if scheme.lower() != "pkg":
        return None, "PURL scheme must be 'pkg'"

    remainder = _strip_leading(remainder, "/")
    raw_type, remainder = _split_once_left(remainder, "/")
    if remainder == None:
        return None, "PURL is missing a name"
    type = raw_type.lower()
    if not type:
        return None, "Mandatory property 'type' not set"

    name_remainder = remainder
    remainder, raw_version = _split_once_right(remainder, "@")
    version = None
    if raw_version != None and "/" in raw_version:
        remainder = name_remainder
        raw_version = None
    if raw_version != None:
        version, err = _decode(raw_version, "version")
        if err:
            return None, err

    namespace, raw_name = _split_once_right(remainder, "/")
    if raw_name == None:
        namespace = None
        raw_name = remainder

    name, err = _decode(raw_name, "name")
    if err:
        return None, err

    namespace_value = None
    if namespace:
        namespace_segments, err = _decode_segments(namespace, "namespace")
        if err:
            return None, err
        namespace_value = "/".join(namespace_segments) if namespace_segments else None

    err = validate(
        type = type,
        namespace = namespace_value,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
        strict = strict,
    )
    if err:
        return None, err

    purl, err = normalize(
        type = type,
        namespace = namespace_value,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
    if err:
        return None, err

    return _as_dict(purl), None
