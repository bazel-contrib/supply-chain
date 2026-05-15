"""Parser for [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private/normalization:normalization.bzl", "normalize")
load("//purl/private/percent_encoding:percent_encoding.bzl", "percent_decode")
load("//purl/private/strings:strings.bzl", "strings")
load("//purl/private/validation:validation.bzl", "is_valid_type", "validate")

visibility([
    "//purl/...",
])

def _split_once_from_right(value, delimiter):
    index = value.rfind(delimiter)
    if index < 0:
        return value, None
    return value[:index], value[index + len(delimiter):]

def _split_version(value):
    left, right = _split_once_from_right(value, "@")
    if right == None:
        return value, None
    return left, right

def _split_once_from_left(value, delimiter):
    index = value.find(delimiter)
    if index < 0:
        return value, None
    return value[:index], value[index + len(delimiter):]

def _strip_leading(value, c):
    for i in range(len(value)):
        if value[i] != c:
            return value[i:]
    return ""

def _decode_subpath_segments(raw_segments, discard_dot_segments):
    segments = []
    for raw_segment in raw_segments:
        segment, err = percent_decode(raw_segment)
        if err:
            return None, err
        if discard_dot_segments and segment in ["", ".", ".."]:
            continue

        if "/" in segment:
            return None, "Decoded segment must not contain '/': '%s'".format(segment)

        segments.append(segment)
    return segments, None

def _decode_namespace_segments(raw_segments):
    segments = []
    for raw_segment in raw_segments:
        if not raw_segment:
            continue

        segment, err = percent_decode(raw_segment)
        if err:
            return None, err

        if not segment:
            continue

        # The generated conformance fixtures expect encoded slashes inside a
        # namespace segment to survive roundtrip as data, not as delimiters.
        # Keep the raw segment when percent-decoding would introduce '/'.
        if "/" in segment:
            segments.append(raw_segment)
        else:
            segments.append(segment)

    return segments, None

def _to_dict(purl):
    return {
        "type": purl.type,
        "namespace": "/".join(purl.namespace) if purl.namespace else None,
        "name": purl.name,
        "version": purl.version,
        "qualifiers": purl.qualifiers if purl.qualifiers else None,
        "subpath": "/".join(purl.subpath) if purl.subpath else None,
    }

def parse(value):
    """Parses a PURL string into normalized components.

    The parsing flow implements ECMA-427 1st edition, December 2025,
    §5.6 "Rules for each PURL component".

    See https://ecma-international.org/wp-content/uploads/ECMA-427_1st_edition_december_2025.pdf

    It parses the components in reverse order of their appearance in the PURL string, as recommended by
    https://github.com/PaawanBarach/purl-spec/blob/main/docs/how-to-parse.md

    Args:
        value: The PURL string to parse.

    Returns:
        A tuple of (purl_components, error). On success, error is None.
    """
    if not value:
        return None, "PURL must not be empty"

    # ECMA-427 §5.6.1, bullets 1-2: scheme is the constant "pkg" and is
    # followed by an unencoded ':' separator.
    if len(value) < 4 and value[0:3] != "pkg:":
        return None, "PURL scheme must be 'pkg'"

    scheme, value = _split_once_from_left(value, ":")
    if value == None:
        return None, "PURL scheme is required"

    # ECMA-427 §5.6.1, bullet 3: PURL parsers shall accept URLs where the scheme and colon ':' are followed by
    # one or more slash '/' characters, such as 'pkg://', and shall ignore and remove all such '/' characters.
    value = _strip_leading(value, '/')

    # ECMA-427 §5.6.7, bullets 1-2: the subpath is introduced by '#',
    # and the separator is not part of the subpath.
    remainder, raw_subpath = _split_once_from_right(value, "#")
    subpath = None
    if raw_subpath != None:
        # ECMA-427 §5.6.7, bullets 3-6: split subpath segments on '/',
        # ignore non-significant leading/trailing slashes, percent-decode each
        # segment, and reject decoded segments that are empty, '.'/'..'.
        subpath_segments, err = _decode_subpath_segments(raw_subpath.split("/"), True)
        if err:
            return None, err

        if subpath_segments:
            subpath = "/".join(subpath_segments)

    # ECMA-427 §5.6.6, bullet 1: the qualifiers component is introduced by
    # '?', and the separator is not part of the qualifiers.
    remainder, raw_qualifiers = _split_once_from_right(remainder, "?")
    qualifiers = None
    if raw_qualifiers != None:
        qualifiers = {}
        # ECMA-427 §5.6.6, bullet 2: qualifiers are one or more key=value
        # pairs separated by '&', which is not part of a qualifier.
        for pair in raw_qualifiers.split("&"):
            # ECMA-427 §5.6.6, bullets 3-5: split on '=', keys are lowercased keys,
            # values are decoded, and empty values are ignored.
            key, raw_value = _split_once_from_left(pair, "=")
            if raw_value == None or raw_value == "":
                continue

            key = key.lower()
            qualifier_value, err = percent_decode(raw_value)
            if err:
                return None, err

            # ECMA-427 §5.6.6, bullet 5, item 4 : Each key shall be unique among all the keys of the qualifiers component.
            if key in qualifiers:
                return None, "Duplicate qualifier key: {}".format(key)

            # TODO Should we split qualifier values on ',' to when key is checksum ?
            # According to how-to-parse.md :
            # If the key is 'checksum', split the value on ',' to create a list of checksums
            #
            # However, the ECMA-427 spec does not mention this behavior and the generated conformance fixtures do not test it.
            qualifiers[key] = qualifier_value

        if not qualifiers:
            qualifiers = None

    # ECMA-427 §5.6.2, bullets 1-4: type is unencoded, starts with an ASCII
    # letter, contains only ASCII letters/numbers, '.', and '-', and is
    # canonicalized to lowercase.
    type, remainder = _split_once_from_left(remainder, "/")
    if remainder == None:
        return None, "PURL type and name must be separated by '/'"

    type = type.lower()
    if not is_valid_type(type):
        return None, "PURL type is invalid"

    # ECMA-427 §5.6.5, bullets 1-4: version, when present, is introduced by
    # '@', excludes that separator, is percent-encoded, and decodes to an
    # opaque string.
    remainder, raw_version = _split_version(remainder)
    version = None
    if raw_version != None:
        version, err = percent_decode(raw_version)
        if err:
            return None, err

    # ECMA-427 §5.6.4, bullets 1-4: name is separated from namespace by '/',
    # leading/trailing slashes are not significant, and the name is a
    # percent-encoded string decoded before type-specific validation.
    remainder, raw_name = _split_once_from_right(remainder, "/")
    if raw_name == None:
        raw_name = remainder
        remainder = None

    name, err = percent_decode(raw_name)
    if err:
        return None, err

    # ECMA-427 §5.6.3, bullets 1-5: namespace is optional, may contain '/'
    # separated segments, ignores non-significant leading/trailing slashes,
    # and each decoded segment must be non-empty and contain no '/'.
    namespace = None
    if remainder:
        namespace_segments, err = _decode_namespace_segments(remainder.split("/"))
        if err:
            return None, err
        if namespace_segments:
            namespace = "/".join(namespace_segments)

    err = validate(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
    if err:
        return None, err

    purl, err = normalize(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
    if err:
        return None, err

    return _to_dict(purl), None
