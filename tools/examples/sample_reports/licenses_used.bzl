"""An example of gathering and processing just license information."""

load(
    "@package_metadata//:defs.bzl",
    "PackageAttributeInfo",
)
load(
    "@package_metadata//licenses/providers:license_kind_info.bzl",
    "LicenseKindInfo",
)
load(
    "@supply_chain_tools//gather_metadata:core.bzl",
    "gather_metadata_info_common",
    "should_traverse",
)
load(
    "@supply_chain_tools//gather_metadata:providers.bzl",
    "null_transitive_metadata_info",
    "TransitiveMetadataInfo",
)

def _gather_licenses_info_impl(target, ctx):
    return gather_metadata_info_common(
        target,
        ctx,
        want_providers = [PackageAttributeInfo, LicenseKindInfo],
        provider_factory = TransitiveMetadataInfo,
        null_provider_instance = null_transitive_metadata_info,
        filter_func = should_traverse,
    )

gather_licenses_info = aspect(
    doc = """Collects metadata providers into a single TransitiveMetadataInfo provider.""",
    implementation = _gather_licenses_info_impl,
    attr_aspects = ["*"],
    attrs = {
        "_trace": attr.label(default = "@supply_chain_tools//gather_metadata:trace_target"),
    },
    provides = [TransitiveMetadataInfo],
    apply_to_generating_rules = True,
)

def _licenses_used_impl(ctx):
    # Gather all licenses and make make a report from that

    # TODO: Replace this
    # The code below just dumps the collected metadata providers in a somewhat
    # pretty printed way.  In reality, we need to read the files associated with
    # each attribute to get the real data. So this should be a rule to pass
    # all the files to a helper which generates a formated report.
    # That is clearly a job for another day.
    out = []
    for dep in ctx.attr.deps:
        if TransitiveMetadataInfo not in dep:
            continue
        t_m_i = dep[TransitiveMetadataInfo]
        out.append("Target: %s\n" % str(t_m_i.target))
        for item in t_m_i.metadata.to_list():
           kind = item.kind if hasattr(item, "kind") else "<unknown>"
           props = ["kind: %s" % kind]
           for field in sorted(dir(item)):
               # skip files because it is a depset of files we need to read.
               if field in ("files", "kind"):
                   continue
               value = getattr(item, field)
               if field == "attributes":
                   props.append("%s: %s" % (field, value.path))
               else:
                   props.append("%s: %s" % (field, value))
           out.append("   %s\n" % ", ".join(props))
    ctx.actions.write(ctx.outputs.out, "".join(out) + "\n")
    return [DefaultInfo(files = depset([ctx.outputs.out]))]

_licenses_used = rule(
    implementation = _licenses_used_impl,
    doc = """Internal tmplementation method for licenses_used().""",
    attrs = {
        "deps": attr.label_list(
            doc = """List of targets to collect LicenseInfo for.""",
            aspects = [gather_licenses_info],
        ),
        "out": attr.output(
            doc = """Output file.""",
            mandatory = True,
        ),
    },
)

def licenses_used(name, deps, out = None, **kwargs):
    """Collects LicensedInfo providers for a set of targets and writes as JSON.

    The output is a single JSON array, with an entry for each license used.
    See gather_licenses_info.bzl:write_licenses_info() for a description of the schema.

    Args:
      name: The target.
      deps: A list of targets to get LicenseInfo for. The output is the union of
            the result, not a list of information for each dependency.
      out: The output file name. Default: <name>.json.
      **kwargs: Other args

    Usage:

      licenses_used(
          name = "license_info",
          deps = [":my_app"],
          out = "license_info.json",
      )
    """
    if not out:
        out = name + ".txt"
    _licenses_used(name = name, deps = deps, out = out, **kwargs)
