"""Analysis tests for PURL type registration through the module extension."""

load("@package_metadata_purl_types//:normalizers.bzl", "TYPE_NORMALIZERS")
load("@package_metadata_purl_types//:validators.bzl", "TYPE_VALIDATORS")
load("@rules_testing//lib:analysis_test.bzl", "analysis_test")

_RegistryAnchorInfo = provider(fields = ["label"])

def _registry_anchor_impl(ctx):
    return [_RegistryAnchorInfo(label = str(ctx.label))]

_registry_anchor = rule(
    implementation = _registry_anchor_impl,
    attrs = {},
    provides = [_RegistryAnchorInfo],
)

def _assert_present(mapping, key, kind):
    if key not in mapping:
        fail("{} registry does not contain '{}'".format(kind, key))

def _assert_normalized(result, expected_type, expected_namespace, expected_name):
    if result.type != expected_type:
        fail("Expected type '{}', got '{}'".format(expected_type, result.type))
    if result.namespace != expected_namespace:
        fail("Expected namespace '{}', got '{}'".format(expected_namespace, result.namespace))
    if result.name != expected_name:
        fail("Expected name '{}', got '{}'".format(expected_name, result.name))

def _builtin_type_registration_test_impl(env, target):
    _assert_present(TYPE_VALIDATORS, "npm", "validator")
    _assert_present(TYPE_NORMALIZERS, "bitbucket", "normalizer")

    err = TYPE_VALIDATORS["bitbucket"](
        type = "bitbucket",
        namespace = ["Workspace"],
        name = "Repo",
        version = "1.0.0",
        qualifiers = None,
        subpath = None,
    )
    if err != None:
        fail("Built-in bitbucket validator rejected a valid package: {}".format(err))

    normalized = TYPE_NORMALIZERS["bitbucket"](
        type = "bitbucket",
        namespace = ["Workspace"],
        name = "Repo",
        version = "1.0.0",
        qualifiers = None,
        subpath = None,
    )
    if normalized.type != "bitbucket":
        fail("Expected type 'bitbucket', got '{}'".format(normalized.type))
    if normalized.namespace != ["workspace"]:
        fail("Expected namespace '['workspace']', got '{}'".format(normalized.namespace))
    if normalized.name != "repo":
        fail("Expected name 'repo', got '{}'".format(normalized.name))

def builtin_type_registration_test(*, name, target):
    _registry_anchor(name = "{}_anchor".format(name))
    analysis_test(
        name = name,
        target = ":{}_anchor".format(name),
        impl = _builtin_type_registration_test_impl,
    )

def _ansible_type_registration_test_impl(env, target):
    _assert_present(TYPE_VALIDATORS, "ansible", "validator")
    _assert_present(TYPE_NORMALIZERS, "ansible", "normalizer")

    err = TYPE_VALIDATORS["ansible"](
        type = "ansible",
        namespace = "Collection",
        name = "Role",
        version = "1.0.0",
        qualifiers = {
            "download_url": "https://example.invalid/role.tgz",
        },
        subpath = None,
    )
    if err != None:
        fail("Custom ansible validator rejected a valid package: {}".format(err))

    normalized = TYPE_NORMALIZERS["ansible"](
        type = "ansible",
        namespace = "Collection",
        name = "Role",
        version = "1.0.0",
        qualifiers = {
            "download_url": "https://example.invalid/role.tgz",
        },
        subpath = None,
    )
    _assert_normalized(normalized, "ansible", "collection", "role")

def ansible_type_registration_test(*, name, target):
    _registry_anchor(name = "{}_anchor".format(name))
    analysis_test(
        name = name,
        target = ":{}_anchor".format(name),
        impl = _ansible_type_registration_test_impl,
    )

def _dependency_apko_type_registration_test_impl(env, target):
    _assert_present(TYPE_VALIDATORS, "apko", "validator")
    _assert_present(TYPE_NORMALIZERS, "apko", "normalizer")

    err = TYPE_VALIDATORS["apko"](
        type = "apko",
        namespace = None,
        name = "image",
        version = "1.0.0",
        qualifiers = None,
        subpath = None,
    )
    if err != None:
        fail("apko validator should come from the root override, but returned: {}".format(err))

def dependency_apko_type_registration_test(*, name, target):
    _registry_anchor(name = "{}_anchor".format(name))
    analysis_test(
        name = name,
        target = ":{}_anchor".format(name),
        impl = _dependency_apko_type_registration_test_impl,
    )

def _apk_type_override_test_impl(env, target):
    _assert_present(TYPE_VALIDATORS, "apk", "validator")
    _assert_present(TYPE_NORMALIZERS, "apk", "normalizer")

    err = TYPE_VALIDATORS["apk"](
        type = "apk",
        namespace = "Alpine",
        name = "BusyBox",
        version = "1.36.1-r0",
        qualifiers = {
            "arch": "x86_64",
        },
        subpath = None,
    )
    if err != None:
        fail("APK override rejected a valid package: {}".format(err))

    err = TYPE_VALIDATORS["apk"](
        type = "apk",
        namespace = "Alpine",
        name = "BusyBox",
        version = "1.36.1-r0",
        qualifiers = {
            "unsupported": "value",
        },
        subpath = None,
    )
    if err == None:
        fail("APK override accepted an unsupported qualifier")

def apk_type_override_test(*, name, target):
    _registry_anchor(name = "{}_anchor".format(name))
    analysis_test(
        name = name,
        target = ":{}_anchor".format(name),
        impl = _apk_type_override_test_impl,
    )

def _root_override_precedence_test_impl(env, target):
    normalized = TYPE_NORMALIZERS["apko"](
        type = "APKO",
        namespace = "Root",
        name = "Image",
        version = "1.0.0",
        qualifiers = None,
        subpath = None,
    )
    _assert_normalized(normalized, "apko", "root", "image")

def root_override_precedence_test(*, name, target):
    _registry_anchor(name = "{}_anchor".format(name))
    analysis_test(
        name = name,
        target = ":{}_anchor".format(name),
        impl = _root_override_precedence_test_impl,
    )
