"""Utils to validate [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private/strings:strings.bzl", "strings")
load("//purl/private/validation:alpm.bzl", "validate_alpm")
load("//purl/private/validation:chrome_extension.bzl", "validate_chrome_extension")
load("//purl/private/validation:cpan.bzl", "validate_cpan")
load("//purl/private/validation:hackage.bzl", "validate_hackage")
load("//purl/private/validation:identity.bzl", "validate_identity")
load("//purl/private/validation:julia.bzl", "validate_julia")
load("//purl/private/validation:otp.bzl", "validate_otp")
load("//purl/private/validation:qualifiers.bzl", "validate_strict_qualifiers")
load("//purl/private/validation:pub.bzl", "validate_pub")
load("//purl/private/validation:pypi.bzl", "validate_pypi")
load("//purl/private/validation:swift.bzl", "validate_swift")
load("//purl/private/validation:vscode_extension.bzl", "validate_vscode_extension")

visibility([
    "//purl/private",
])

_validators = {
    "alpm": validate_alpm,
    "apk": validate_identity,
    "bazel": validate_identity,
    "bitbucket": validate_identity,
    "bitnami": validate_identity,
    "cargo": validate_identity,
    "chrome-extension": validate_chrome_extension,
    "cocoapods": validate_identity,
    "composer": validate_identity,
    "conan": validate_identity,
    "conda": validate_identity,
    "cpan": validate_cpan,
    "cran": validate_identity,
    "deb": validate_identity,
    "docker": validate_identity,
    "gem": validate_identity,
    "generic": validate_identity,
    "github": validate_identity,
    "golang": validate_identity,
    "hackage": validate_hackage,
    "hex": validate_identity,
    "huggingface": validate_identity,
    "julia": validate_julia,
    "luarocks": validate_identity,
    "maven": validate_identity,
    "mlflow": validate_identity,
    "npm": validate_identity,
    "nuget": validate_identity,
    "oci": validate_identity,
    "opam": validate_identity,
    "otp": validate_otp,
    "pub": validate_pub,
    "pypi": validate_pypi,
    "qpkg": validate_identity,
    "rpm": validate_identity,
    "swid": validate_identity,
    "swift": validate_swift,
    "vscode-extension": validate_vscode_extension,
    "yocto": validate_identity,
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

    if strict:
        err = validate_strict_qualifiers(type, qualifiers)
        if err:
            return err

    validator = _validators.get(type.lower())
    if not validator:
        return None

    return validator(
        type = type,
        namespace = namespace,
        name = name,
        version = version,
        qualifiers = qualifiers,
        subpath = subpath,
        strict = strict,
    )
