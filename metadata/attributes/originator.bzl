"""Declares the package originator attribute."""

load(":agent.bzl", "agent_attribute", "agent_rule")

visibility("public")

KIND = "build.bazel.attribute.originator"
_originator = agent_rule(KIND)

def originator(
        *,
        name,
        entity_name,
        type = "organization",
        url = None,
        email = None,
        tags = None,
        visibility = None):
    """Declares who originally created a package.

    The originator is the entity that originally created the package; it may
    differ from the supplier that distributes it. This is typically passed to
    the `attributes` of a `package_metadata` target.

    Usage:

    ```starlark
    load("@package_metadata//attributes:originator.bzl", "originator")

    originator(
        name = "originator",
        entity_name = "A. Creator",
        type = "person",
        email = "creator@example",
    )
    ```

    Format: the `PackageAttributeInfo` kind is
    `build.bazel.attribute.originator`; its JSON object has `name` and `type`,
    and includes `url` and `email` only when set. CycloneDX maps this to the
    component publisher string, so type, URL, and email are not represented.
    SPDX 2.3 maps it to `PackageOriginator`, preserving type and email (in
    parentheses) but not URL. SPDX 3.0.1 maps it to a `Person` or
    `Organization` referenced by the package's `originatedBy`, preserving
    email as an email external identifier and URL as an `altWebPage` external
    reference.

    Args:
      name: A unique name for this target.
      entity_name: Required. UTF-8 name of the original creator.
      type: `organization` or `person`; defaults to `organization`.
      url: Optional URL for the originator.
      email: Optional email address for the originator.
      tags: A list of arbitrary tags to apply to this target.
      visibility: The visibility of this target.
    """
    agent_attribute(
        rule = _originator,
        name = name,
        entity_name = entity_name,
        type = type,
        url = url,
        email = email,
        tags = tags,
        visibility = visibility,
    )
