"""Validation for chrome-extension PURLs.
"""

load("//purl/private/strings:strings.bzl", "strings")
load("//purl/types:helpers.bzl", "validate_with_specific")

visibility("public")

def _validate_chrome_extension_specific(*, type, namespace, name, version, qualifiers, subpath):
    # https://github.com/package-url/purl-spec/blob/d4710aaa445aea991383385251c166c1bb26f9ba/types/chrome-extension-definition.json#L13
    if namespace:
        return "Chrome extension PURLs must not have a namespace"

    # https://github.com/package-url/purl-spec/blob/d4710aaa445aea991383385251c166c1bb26f9ba/types/chrome-extension-definition.json#L19
    if len(name) != 32:
        return "Chrome extension IDs must be 32 characters"

    for c in name.elems():
        if c < "a" or c > "z":
            return "Chrome extension IDs may only contain characters a-z"

    if version:
        # https://github.com/package-url/purl-spec/blob/d4710aaa445aea991383385251c166c1bb26f9ba/types/chrome-extension-definition.json#L26
        segments = version.split(".")
        if len(segments) > 4:
            return "Chrome extension versions may have at most four segments"
        for segment in segments:
            if not segment:
                return "Chrome extension version segments must not be empty"
            for c in strings.bytes.from_string(segment):
                if not strings.ascii.is_alphanumeric(c) or strings.ascii.is_alpha(c):
                    return "Chrome extension version segments must be numeric"

    return None

def validate_chrome_extension(*, type, namespace, name, version, qualifiers, subpath):
    return validate_with_specific(type, _validate_chrome_extension_specific, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)
