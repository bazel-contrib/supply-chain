load(":registration_test.bzl", "ansible_type_registration_test", "apk_type_override_test", "builtin_type_registration_test", "dependency_apko_type_registration_test", "root_override_precedence_test")

builtin_type_registration_test(
    name = "builtin_type_registration_test",
    target = "@package_metadata_purl_types//:validators.bzl",
)

ansible_type_registration_test(
    name = "ansible_type_registration_test",
    target = "@purl_ansible_type_test//:ansible_type.bzl",
)

dependency_apko_type_registration_test(
    name = "dependency_apko_type_registration_test",
    target = "@rules_apko_test//:dependency_apko_type.bzl",
)

apk_type_override_test(
    name = "apk_type_override_test",
    target = "@package_metadata_purl_types//:validators.bzl",
)

root_override_precedence_test(
    name = "root_override_precedence_test",
    target = "@package_metadata_purl_types//:normalizers.bzl",
)
