"""A second module with an identically named image-layer rule."""

def _image_layer_impl(ctx):
    return [DefaultInfo()]

image_layer = rule(
    implementation = _image_layer_impl,
    attrs = {
        "srcs": attr.label_list(),
        "other_provider_marker": attr.string(),
        "applicable_licenses": attr.label_list(),
        "_supply_chain_relationships": attr.string_dict(default = {
            "srcs": "contains",
        }),
    },
)
