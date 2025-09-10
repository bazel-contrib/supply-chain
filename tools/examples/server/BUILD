# Examples of applications and interactions with licenses

load("@rules_cc//cc:cc_binary.bzl", "cc_binary")

# load("@supply_chain_tools//rules:compliance.bzl", "check_license")
load("@supply_chain_tools//examples/sample_reports:licenses_used.bzl", "licenses_used")
load("@supply_chain_tools//examples/sbom:sbom.bzl", "sbom")
load("@supply_chain_tools//examples/vendor/constant_gen:defs.bzl", "constant_gen")

package(
    default_package_metadata = ["//:package_metadata"],
    default_visibility = ["//examples:__subpackages__"],
)

# Sample
constant_gen(
    name = "message",
    text = "Hello, world.",
    var = "server_message",
)

# This server is clean to ship.
cc_binary(
    name = "my_server",
    srcs = ["server.cc"],
    deps = [":message"],
)

# Verify the licenses are what we expect. The output shows that
# :my_server only uses the unencumbered license type.
licenses_used(
    name = "server_licenses",
    out = "server_licenses.txt",
    deps = [":my_server"],
)

# This server uses something under a restricted license.
cc_binary(
    name = "my_violating_server",
    srcs = ["server.cc"],
    deps = [
        ":message",
        "//examples/lib:private_lib",
	"//examples/lib:lib_with_third_party_deps",
        "@supply_chain_tools//examples/vendor/acme",
        "@supply_chain_tools//examples/vendor/libhhgttg",
    ],
)

licenses_used(
    name = "violating_server_licenses",
    out = "violating_server_licenses.txt",
    deps = [":my_violating_server"],
)

sbom(
    name = "server_sbom",
    out = "server_sbom.txt",
    target = ":my_violating_server",
)

# TODO(aiuto): Turn this strictly into a compliance test.
#check_license(
#    name = "check_server",
#    check_conditions = False,
#    license_texts = "server_licenses.txt",
#    report = "server_report.txt",
#    deps = [
#        ":my_server",
#    ],
#)

#py_test(
#    name = "server_licenses_test",
#    srcs = ["server_licenses_test.py"],
#    data = [":server_licenses.json"],
#    python_version = "PY3",
#    deps = [
#        "@supply_chain_tools//tests:license_test_utils",
#    ],
#)
