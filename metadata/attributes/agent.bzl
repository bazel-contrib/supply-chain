"""Shared implementation for package supplier and originator attributes."""

load("//providers:package_attribute_info.bzl", "PackageAttributeInfo")

def _agent_impl(kind):
    def impl(ctx):
        attributes = {
            "name": ctx.attr.entity_name,
            "type": ctx.attr.type,
        }
        if ctx.attr.url:
            attributes["url"] = ctx.attr.url
        if ctx.attr.email:
            attributes["email"] = ctx.attr.email

        output = ctx.actions.declare_file("{}.package-attribute.json".format(ctx.attr.name))
        ctx.actions.write(output = output, content = json.encode(attributes))
        return [
            DefaultInfo(files = depset(direct = [output])),
            PackageAttributeInfo(kind = kind, attributes = output),
        ]

    return impl

def agent_rule(kind):
    """Creates the private rule used by supplier and originator."""
    return rule(
        implementation = _agent_impl(kind),
        attrs = {
            "entity_name": attr.string(mandatory = True),
            "type": attr.string(
                default = "organization",
                values = ["organization", "person"],
            ),
            "url": attr.string(),
            "email": attr.string(),
        },
    )

def agent_attribute(
        rule,
        name,
        entity_name,
        type,
        url,
        email,
        tags,
        visibility):
    """Invokes an agent rule after validating its public enum argument."""
    if type not in ["organization", "person"]:
        fail("type must be one of 'organization' or 'person', got {!r}".format(type))
    rule(
        name = name,
        entity_name = entity_name,
        type = type,
        url = url or "",
        email = email or "",
        tags = tags,
        visibility = visibility,
        applicable_licenses = [],
    )
