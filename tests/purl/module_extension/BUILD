load("@package_metadata_purl_types//:defs.bzl", "PURL_TYPE_OVERRIDES")
load("//purl/module_extension:apk_override.bzl", "apk_type")
load("//purl/module_extension:ansible_type_registration_test.bzl", "ansible_type_registration_test")
load("//purl/module_extension:builtin_type_registration_test.bzl", "builtin_type_registration_test")
load("//purl/module_extension:purl_type_override_test.bzl", "purl_type_override_test")
load("//purl/module_extension:root_apko_override.bzl", "root_apko_type")
load("//purl/module_extension:root_override_precedence_test.bzl", "root_override_precedence_test")

apk_type(
    name = "apk_type",
    visibility = ["//visibility:public"],
)

root_apko_type(
    name = "root_apko_type",
    visibility = ["//visibility:public"],
)

purl_type_override_test(
    name = "apk_type_override_test",
    target = PURL_TYPE_OVERRIDES["apk"],
)

ansible_type_registration_test(
    name = "ansible_type_registration_test",
    target = PURL_TYPE_OVERRIDES["ansible"],
)

root_override_precedence_test(
    name = "root_override_precedence_test",
    target = PURL_TYPE_OVERRIDES["apko"],
)

builtin_type_registration_test(
    name = "builtin_type_registration_test",
    target = PURL_TYPE_OVERRIDES["npm"],
)
