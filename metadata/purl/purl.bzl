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

purl = struct(
    bazel = _bazel,
    builder = builder,
)
