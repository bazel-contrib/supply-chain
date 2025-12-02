"""Module defining urils for [purl](https://github.com/package-url/purl-spec)s."""

visibility("public")

_DEFAULT_REGISTRY = "https://bcr.bazel.build"

def _bazel(name, version, registry = _DEFAULT_REGISTRY):
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
        registry (str): The URL of the registry that hosts the Bazel module. Defaults to
                        https://bcr.bazel.build.

    Returns:
        The `purl` for the Bazel module (e.g. `pkg:bazel/foo` or
        `pkg:bazel/bar@1.2.3`).
    """

    version_part = "@{}".format(version) if version else ""
    normalized_registry = registry.rstrip("/")
    registry_part = "?repository_url={}".format(normalized_registry) if normalized_registry != _DEFAULT_REGISTRY else ""
    return "pkg:bazel/{}{}{}".format(name, version_part, registry_part)

purl = struct(
    bazel = _bazel,
)
