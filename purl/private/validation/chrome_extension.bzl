"""Validation for Chrome Extension PURLs."""

visibility([
    "//private/validation/...",
])

def _is_chrome_id_char(c):
    return c >= "a" and c <= "p"

def _is_digit(c):
    return c >= "0" and c <= "9"

def validate_chrome_extension(*, type, namespace, name, version, qualifiers, subpath):
    if len(name) != 32:
        return "Chrome extension names must be 32 characters"
    for c in name.elems():
        if not _is_chrome_id_char(c):
            return "Chrome extension names must use characters a-p"

    if version:
        segments = version.split(".")
        if len(segments) > 4:
            return "Chrome extension versions must have at most four numeric segments"
        for segment in segments:
            if not segment:
                return "Chrome extension versions must not have empty segments"
            for c in segment.elems():
                if not _is_digit(c):
                    return "Chrome extension versions must be numeric"

    return None
