"""Transitions for building image manifests with distinct libc selections."""

load("@rules_img//img:providers.bzl", "ImageManifestInfo")

_libc_flag = str(Label("//:libc"))
_python_libc_flag = str(Label("@rules_python//python/config_settings:py_linux_libc"))

def _libc_variant_transition_impl(settings, attr):
    libc = settings[_libc_flag] if attr.libc == "inherit" else attr.libc
    return {
        _libc_flag: libc,
        _python_libc_flag: libc,
    }

_libc_variant_transition = transition(
    implementation = _libc_variant_transition_impl,
    inputs = [_libc_flag],
    outputs = [_libc_flag, _python_libc_flag],
)

def _libc_variant_impl(ctx):
    target = ctx.attr.target[0]
    return [
        target[ImageManifestInfo],
        DefaultInfo(files = target[DefaultInfo].files),
    ]

libc_variant = rule(
    implementation = _libc_variant_impl,
    attrs = {
        "libc": attr.string(
            default = "inherit",
            values = ["inherit", "glibc", "musl"],
        ),
        "target": attr.label(
            cfg = _libc_variant_transition,
            providers = [ImageManifestInfo],
        ),
    },
)
