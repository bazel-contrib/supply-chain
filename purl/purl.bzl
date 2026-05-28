"""Module defining urils for [purl](https://github.com/package-url/purl-spec)s."""

load("//private:builder.bzl", _builder = "builder")
load("//private:parser.bzl", _parse = "parse")

visibility("public")

_DEFAULT_REGISTRY = "https://bcr.bazel.build"

def bazel(name, version, registry = _DEFAULT_REGISTRY):
    """Defines a `purl` for a Bazel module.

    This is typically used to construct `purl` for `package_metadata` targets in
    Bazel modules.

    This is **NOT** supported in `WORKSPACE` mode.

    Example:

    ```starlark
    load("@purl.bzl", "purl")

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
        registry (str): The URL of the registry that hosts the Bazel module. Defaults to
                        https://bcr.bazel.build.

    Returns:
        The `purl` for the Bazel module (e.g. `pkg:bazel/foo` or
        `pkg:bazel/bar@1.2.3`).
    """

    builder = _builder().type("bazel").name(name).version(version)
    if registry != _DEFAULT_REGISTRY:
        builder = builder.qualifier("repository_url", registry)

    return builder.build()

builder = _builder
parse = _parse
