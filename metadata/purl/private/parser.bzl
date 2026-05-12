"""Parser for [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private/normalization:normalization.bzl", "normalize")
load("//purl/private/percent_encoding:percent_encoding.bzl", "percent_decode")
load("//purl/private/strings:strings.bzl", "strings")
load("//purl/private/validation:validation.bzl", "validate")

visibility([
    "//purl/...",
])

def _split_once_from_right(value, delimiter):
    index = value.rfind(delimiter)
    if index < 0:
        return value, None
    return value[:index], value[index + len(delimiter):]

def _split_version(value, strict):
    left, right = _split_once_from_right(value, "@")
    if right == None:
        return value, None
    if (not strict) and ("/" in right):
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

def _strip_trailing(value, c):
    for i in range(len(value)):
        index = len(value) - i - 1
        if value[index] != c:
            return value[:index + 1]
    return ""

def _decode_segments(raw_segments, discard_dot_segments, strict):
    segments = []
    for raw_segment in raw_segments:
        segment, err = percent_decode(raw_segment, strict = strict)
        if err:
            return None, err
        if discard_dot_segments and segment in ["", ".", ".."]:
            continue
        if "/" in segment:
            return None, "Path segment contains '/' after percent-decoding"
        segments.append(segment)
    return segments, None

def _is_valid_type(type):
    if not type:
        return False

    first = strings.bytes.from_string(type[0])[0]
    if not strings.ascii.is_alpha(first):
        return False

    for c in strings.bytes.from_string(type):
        if strings.ascii.is_alphanumeric(c):
            continue
        if c in [43, 45, 46]:  # '+', '-', '.'
            continue
        return False

    return True

def _to_dict(purl):
    return {
        "type": purl.type,
        "namespace": "/".join(purl.namespace) if purl.namespace else None,
        "name": purl.name,
        "version": purl.version,
        "qualifiers": purl.qualifiers if purl.qualifiers else None,
        "subpath": "/".join(purl.subpath) if purl.subpath else None,
    }

def parse(value, strict = False):
    """Parses a PURL string into normalized components.

    The parsing flow implements ECMA-427 1st edition, December 2025,
    §5.6 "Rules for each PURL component".

    Args:
        value: The PURL string to parse.
        strict: Whether malformed percent escapes and unknown qualifier keys
                should fail parsing.

    Returns:
        A tuple of (purl_components, error). On success, error is None.
    """

    if not value:
        return None, "PURL must not be empty"

    # ECMA-427 §5.6.7, bullets 1-2: the subpath is introduced by '#',
    # and the separator is not part of the subpath.
    remainder, raw_subpath = _split_once_from_right(value, "#")
    subpath = None
    if raw_subpath != None:
        # ECMA-427 §5.6.7, bullets 3-6: split subpath segments on '/',
        # ignore non-significant leading/trailing slashes, percent-decode each
        # segment, and reject decoded segments that are empty, '.'/'..', or
        # contain '/'.
        subpath_segments, err = _decode_segments(raw_subpath.split("/"), True, strict)
        if err:
            return None, err
        if subpath_segments:
            subpath = "/".join(subpath_segments)

    # ECMA-427 §5.6.6, bullet 1: the qualifiers component is introduced by
    # '?', and the separator is not part of the qualifiers component.
    remainder, raw_qualifiers = _split_once_from_right(remainder, "?")
    qualifiers = None
    if raw_qualifiers != None:
        qualifiers = {}
        # ECMA-427 §5.6.6, bullet 2: qualifiers are one or more key=value
        # pairs separated by '&', which is not part of a qualifier.
        for pair in raw_qualifiers.split("&"):
            key, raw_value = _split_once_from_left(pair, "=")
            if raw_value == None:
                raw_value = ""

            # ECMA-427 §5.6.6, bullets 3-5: split on '=', lowercase keys
            # are validated later, values are decoded, and empty values are
            # treated as if the key=value pair did not exist.
            key = key.lower()
            qualifier_value, err = percent_decode(raw_value, strict = strict)
            if err:
                return None, err
            if qualifier_value == "":
                continue

            # ECMA-427 §5.6.6, bullet 6: key syntax, uniqueness, and strict
            # qualifier allow-list checks are enforced by validate().
            qualifiers[key] = qualifier_value

        if not qualifiers:
            qualifiers = None

    # ECMA-427 §5.6.1, bullets 1-2: scheme is the constant "pkg" and is
    # followed by an unencoded ':' separator.
    scheme, remainder = _split_once_from_left(remainder, ":")
    if remainder == None:
        return None, "PURL scheme is required"
    if scheme.lower() != "pkg":
        return None, "PURL scheme must be 'pkg'"

    # ECMA-427 §5.6.1, bullet 3: parsers accept and remove one or more '/'
    # characters following "pkg:" before reading the type.
    remainder = _strip_leading(remainder, "/")

    # ECMA-427 §5.6.2, bullets 1-4: type is unencoded, starts with an ASCII
    # letter, contains only ASCII letters/numbers plus '+', '.', and '-', and is
    # canonicalized to lowercase.
    type, remainder = _split_once_from_left(remainder, "/")
    if remainder == None:
        return None, "PURL type and name must be separated by '/'"
    type = type.lower()
    if not _is_valid_type(type):
        return None, "PURL type is invalid"

    # ECMA-427 §5.6.5, bullets 1-4: version, when present, is introduced by
    # '@', excludes that separator, is percent-encoded, and decodes to an
    # opaque string.
    remainder, raw_version = _split_version(remainder, strict)
    version = None
    if raw_version != None:
        version, err = percent_decode(raw_version, strict = strict)
        if err:
            return None, err

    # ECMA-427 §5.6.4, bullets 1-4: name is separated from namespace by '/',
    # leading/trailing slashes are not significant, and the name is a
    # percent-encoded string decoded before type-specific validation.
    remainder = _strip_trailing(remainder, "/")
    remainder, raw_name = _split_once_from_right(remainder, "/")
    if raw_name == None:
        raw_name = remainder
        remainder = ""
    name, err = percent_decode(raw_name, strict = strict)
    if err:
        return None, err

    # ECMA-427 §5.6.3, bullets 1-5: namespace is optional, may contain '/'
    # separated segments, ignores non-significant leading/trailing slashes,
    # and each decoded segment must be non-empty and contain no '/'.
    namespace = None
    if remainder:
        namespace_segments, err = _decode_segments([s for s in remainder.split("/") if s], False, strict)
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
        strict = strict,
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
