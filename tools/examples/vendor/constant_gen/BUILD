# An example of a code generator with a distinct license for the generated code.

load("@package_metadata//licenses/rules:license.bzl", "license")
load("@supply_chain_tools//examples/sample_reports:licenses_used.bzl", "licenses_used")
load(":defs.bzl", "constant_gen")

package(
    default_package_metadata = [":license_of_the_package"],
    default_visibility = ["//examples:__subpackages__"],
)

# The default license for an entire package is typically named "license".
license(
    name = "license_of_the_package",
    kind = "@supply_chain_tools//examples/my_org/licenses:generic_restricted",
    text = "LICENSE",
)

license(
    name = "license_for_emitted_code",
    kind = "@supply_chain_tools//examples/my_org/licenses:unencumbered",
    text = "LICENSE_OF_OUTPUT",
)

# The generator itself will be licensed under :license
py_binary(
    name = "constant_generator",
    srcs = ["constant_generator.py"],
    python_version = "PY3",
)

# Sample: This target will be licensed under :license_for_emitted_code
constant_gen(
    name = "libhello",
    text = "Hello, world.",
    var = "hello_world",
)

# Verify the licenses are what we expect
licenses_used(
    name = "generator_licenses",
    out = "generator_licenses.txt",
    deps = [":constant_generator"],
)

licenses_used(
    name = "generated_code_licenses",
    # Note: using default output file name
    deps = [":libhello"],
)

# TODO(aiuto): Add a test that shows what we expect
# For now:
# bazel build //examples/examples/vendor/constant_gen:all
# examine bazel-bin/examples/vendor/constant_gen/generator_licenses.json
# examine bazel-bin/examples/vendor/constant_gen/generated_code_licenses.json
