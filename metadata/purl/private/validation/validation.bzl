"""Utils to validate [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private:type_definitions.bzl", "TYPE_DEFINITIONS")
load("//purl/private/strings:strings.bzl", "strings")

visibility([
    "//purl/private",
])

def _has_value(value):
    return value != None and value != "" and value != []

def _is_lower_alpha(c):
    return c >= 97 and c <= 122

def _is_digit(c):
    return c >= 48 and c <= 57

def _matches_chrome_extension_id(value):
    if len(value) != 32:
        return False
    for c in strings.bytes.from_string(value):
        if not _is_lower_alpha(c):
            return False
    return True

def _matches_chrome_extension_version(value):
    segments = value.split(".")
    if len(segments) < 1 or len(segments) > 4:
        return False
    for segment in segments:
        if not segment:
            return False
        for c in strings.bytes.from_string(segment):
            if not _is_digit(c):
                return False
    return True

def _matches_pub_name(value):
    for c in strings.bytes.from_string(value):
        if _is_lower_alpha(c) or _is_digit(c) or c == 95:
            continue
        return False
    return True

def _matches_pattern(pattern, value):
    if pattern == "chrome_extension_id":
        return _matches_chrome_extension_id(value)
    if pattern == "chrome_extension_version":
        return _matches_chrome_extension_version(value)
    if pattern == "pub_name":
        return _matches_pub_name(value)
    return True

def _validate_type_definition(type, definition, namespace, name, version, qualifiers, subpath):
    namespace_requirement = definition.get("namespace")
    if namespace_requirement == "required" and not _has_value(namespace):
        return "{} PURLs require a namespace".format(type)
    if namespace_requirement == "prohibited" and _has_value(namespace):
        return "{} PURLs must not have a namespace".format(type)

    if not _has_value(name):
        return "Mandatory property 'name' not set"

    pattern = definition.get("name_pattern")
    if pattern:
        name_value = name.lower() if definition.get("lower_name") else name
        if not _matches_pattern(pattern, name_value):
            return "{} name does not match {}".format(type, pattern)

    pattern = definition.get("version_pattern")
    if pattern and _has_value(version):
        version_value = version.lower() if definition.get("lower_version") else version
        if not _matches_pattern(pattern, version_value):
            return "{} version does not match {}".format(type, pattern)

    for key in definition.get("required_qualifiers", []):
        if not qualifiers or not _has_value(qualifiers.get(key)):
            return "{} PURLs require a '{}' qualifier".format(type, key)

    if definition.get("cpan"):
        if namespace != namespace.upper():
            return "CPAN PURL namespace (author) must be uppercase"
        if "::" in name:
            return "CPAN PURL name must be a distribution name, not a module name (contains '::')"

    return None

def _validate_defined_type(type, *, namespace, name, version, qualifiers, subpath):
    return _validate_type_definition(
        type,
        TYPE_DEFINITIONS[type],
        namespace,
        name,
        version,
        qualifiers,
        subpath,
    )

def _validate_alpm(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_apk(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_bazel(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_bitbucket(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_bitnami(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_cargo(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_chrome_extension(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_cocoapods(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_composer(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_conan(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_conda(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_cpan(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_cran(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_deb(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_docker(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_gem(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_generic(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_github(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_golang(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_hackage(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_hex(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_huggingface(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_julia(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_luarocks(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_maven(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_mlflow(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_npm(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_nuget(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_oci(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_opam(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_otp(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_pub(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_pypi(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_qpkg(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_rpm(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_swid(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_swift(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_vscode_extension(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

def _validate_yocto(*, type, namespace, name, version, qualifiers, subpath):
    return _validate_defined_type(type, namespace = namespace, name = name, version = version, qualifiers = qualifiers, subpath = subpath)

_validators = {
    "alpm": _validate_alpm,
    "apk": _validate_apk,
    "bazel": _validate_bazel,
    "bitbucket": _validate_bitbucket,
    "bitnami": _validate_bitnami,
    "cargo": _validate_cargo,
    "chrome-extension": _validate_chrome_extension,
    "cocoapods": _validate_cocoapods,
    "composer": _validate_composer,
    "conan": _validate_conan,
    "conda": _validate_conda,
    "cpan": _validate_cpan,
    "cran": _validate_cran,
    "deb": _validate_deb,
    "docker": _validate_docker,
    "gem": _validate_gem,
    "generic": _validate_generic,
    "github": _validate_github,
    "golang": _validate_golang,
    "hackage": _validate_hackage,
    "hex": _validate_hex,
    "huggingface": _validate_huggingface,
    "julia": _validate_julia,
    "luarocks": _validate_luarocks,
    "maven": _validate_maven,
    "mlflow": _validate_mlflow,
    "npm": _validate_npm,
    "nuget": _validate_nuget,
    "oci": _validate_oci,
    "opam": _validate_opam,
    "otp": _validate_otp,
    "pub": _validate_pub,
    "pypi": _validate_pypi,
    "qpkg": _validate_qpkg,
    "rpm": _validate_rpm,
    "swid": _validate_swid,
    "swift": _validate_swift,
    "vscode-extension": _validate_vscode_extension,
    "yocto": _validate_yocto,
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

    validator = _validators.get(type)
    if not validator:
        return "Unknown PURL type {}".format(type) if strict else None

    return validator(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
    )
