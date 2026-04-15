"""Module defining urils for [purl](https://github.com/package-url/purl-spec)s."""

load("//purl/private:builder.bzl", "builder")

visibility("public")

def _bazel(name, version):
    """Defines a `purl` for a Bazel module.

    This is typically used to construct `purl` for `package_metadata` targets in
    Bazel modules.

    This is **NOT** supported in `WORKSPACE` mode.

    Example:

    ```starlark
    load("@package_metadata//purl:purl.bzl", "purl")

    package_metadata(
        name = "package_metadata",
        purl = purl.bazel(module_name(), module_version()),
        attributes = [
            # ...
        ],
        visibility = ["//visibility:public"],
    )
    ```

    Args:
        name (str): The name of the Bazel module. Typically
                    [module_name()](https://bazel.build/rules/lib/globals/build#module_name).
        version (str): The version of the Bazel module. Typically
                       [module_version()](https://bazel.build/rules/lib/globals/build#module_version).
                       May be empty or `None`.

    Returns:
        The `purl` for the Bazel module (e.g. `pkg:bazel/foo` or
        `pkg:bazel/bar@1.2.3`).
    """

    return builder().type("bazel").name(name).version(version).build()

def _builder():
    """Creates a fluent builder for constructing Package URLs (PURLs).

    The builder provides a chainable interface for constructing PURLs according to
    the [Package URL specification](https://github.com/package-url/purl-spec).

    The `type` and `name` fields are required. All components are validated and
    normalized according to the PURL spec. Components are automatically percent-encoded
    where necessary, and qualifiers are sorted lexicographically in the output.

    Example - Simple PURL:

        load("@package_metadata//purl:purl.bzl", "purl")

        my_purl = (purl.builder()
            .type("npm")
            .name("lodash")
            .version("4.17.21")
            .build())
        # Result: pkg:npm/lodash@4.17.21

    Example - PURL with namespace, qualifiers, and subpath:

        load("@package_metadata//purl:purl.bzl", "purl")

        my_purl = (purl.builder()
            .type("maven")
            .namespace("org.apache.commons")
            .name("commons-lang3")
            .version("3.12.0")
            .add_qualifier("classifier", "sources")
            .add_qualifier("type", "jar")
            .subpath(["src", "main"])
            .build())
        # Result: pkg:maven/org.apache.commons/commons-lang3@3.12.0?classifier=sources&type=jar#src/main

    Example - Multi-segment namespaces:

        load("@package_metadata//purl:purl.bzl", "purl")

        my_purl = (purl.builder()
            .type("golang")
            .namespace(["github.com", "user", "project"])
            .name("package")
            .build())
        # Result: pkg:golang/github.com/user/project/package

    Returns:
        A builder object with chainable methods:

        - `type(type_name)`: Sets the package type (required). Must be lowercase ASCII.
        - `namespace(namespace)`: Sets the namespace (optional). String or list of strings.
        - `name(name)`: Sets the package name (required).
        - `version(version)`: Sets the package version (optional).
        - `add_qualifier(name, value)`: Adds a qualifier (optional, repeatable).
          Key must start with ASCII letter and contain only lowercase letters,
          numbers, '.', '-', '_'.
        - `subpath(subpath)`: Sets the subpath (optional). List of path segments.
        - `build()`: Constructs the final PURL string. Fails on validation errors.
    """
    return builder()

purl = struct(
    bazel = _bazel,
    builder = _builder,
)
