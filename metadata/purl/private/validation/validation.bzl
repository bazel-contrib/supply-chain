"""Utils to validate [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/types:requirements.bzl", "TYPE_REQUIREMENTS")
load("//purl/private/strings:strings.bzl", "strings")
load("@package_metadata_purl_types//:validators.bzl", "TYPE_VALIDATORS")

visibility([
    "//purl/private",
    "//purl/types",
])

_COMMON_QUALIFIERS = {
    "checksum": True,
    "download_url": True,
    "file_name": True,
    "repository_url": True,
    "vcs_url": True,
    "vers": True,
}

def validate(
        *,
        type = None,
        namespace = None,
        name = None,
        version = None,
        qualifiers = {},
        subpath = None,
        strict = True):
    # Spec §5: Validate required fields are present.
    if not type:
        return "Mandatory property 'type' not set"
    if not name:
        return "Mandatory property 'name' not set"

    type = type.lower()

    for i, c in enumerate(strings.bytes.from_string(type)):
        if strings.ascii.is_alphanumeric(c):
            if i == 0 and not strings.ascii.is_alpha(c):
                return "PURL type must start with an ASCII letter"
            continue
        elif c == 46 or c == 45 or c == 95:  # . - _
            continue

        return "PURL type {} contains illegal character {}".format(type, c)

    if qualifiers:
        for key, value in qualifiers.items():
            # 5.6.6

            if len(key) < 1:
                return "Qualifier key must not be empty string"

            # The key shall be composed only of lowercase ASCII letters and numbers,
            # period '.', dash '-' and underscore '_'.
            for c in strings.bytes.from_string(key):
                if strings.ascii.is_alphanumeric(c):
                    continue
                elif c == 46:  # .
                    continue
                elif c == 45:  # -
                    continue
                elif c == 95:  # _
                    continue

                return "Qualifier key {} contains illegal character {}".format(key, c)

            # A key shall start with an ASCII letter.
            for c in strings.bytes.from_string(key[0]):
                if strings.ascii.is_alpha(c):
                    continue

                return "Qualifier key {} does not start with ASCII letter, got {}".format(key, c)

    validator = TYPE_VALIDATORS.get(type)
    if not validator:
        return "Unknown PURL type {}".format(type) if strict else None

    if strict and qualifiers:
        definition = TYPE_REQUIREMENTS.get(type, {})
        allowed_qualifiers = {
            key: True
            for key in definition.get("qualifiers", [])
        }
        for key in qualifiers.keys():
            key = key.lower()
            if key not in _COMMON_QUALIFIERS and key not in allowed_qualifiers:
                return "{} qualifier '{}' is not supported".format(type, key)

    return validator(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
