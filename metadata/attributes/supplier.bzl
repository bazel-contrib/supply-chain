"""Declares the package supplier attribute."""

load(":agent.bzl", "agent_attribute", "agent_rule")

visibility("public")

KIND = "build.bazel.attribute.supplier"
_supplier = agent_rule(KIND)

def supplier(
        *,
        name,
        entity_name,
        type = "organization",
        url = None,
        email = None,
        tags = None,
        visibility = None):
    """Declares who distributes a package.

    The supplier is the entity that distributes the package, which can differ
    from the entity that originally created it. This is typically passed to
    the `attributes` of a `package_metadata` target.

    Usage:

    ```starlark
    load("@package_metadata//attributes:supplier.bzl", "supplier")

    supplier(
        name = "supplier",
        entity_name = "ACME Corp",
        url = "https://acme.example",
        email = "oss@acme.example",
    )
    ```

    Format: the `PackageAttributeInfo` kind is
    `build.bazel.attribute.supplier`; its JSON object has `name` and `type`,
    and includes `url` and `email` only when set. CycloneDX maps this to the
    component supplier entity, always as an organization even when `type` is
    `person`; URL and email are preserved. SPDX 2.3 maps it to
    `PackageSupplier`, preserving type and email (in parentheses) but not URL.
    SPDX 3.0.1 maps it to a `Person` or `Organization` referenced by the
    package's `suppliedBy`, preserving email as an email external identifier
    and URL as an `altWebPage` external reference.

    Args:
      name: A unique name for this target.
      entity_name: Required. UTF-8 name of the package distributor.
      type: `organization` or `person`; defaults to `organization`.
      url: Optional URL for the supplier.
      email: Optional email address for the supplier.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """
    agent_attribute(
        rule = _supplier,
        name = name,
        entity_name = entity_name,
        type = type,
        url = url,
        email = email,
        tags = tags,
        visibility = visibility,
    )
