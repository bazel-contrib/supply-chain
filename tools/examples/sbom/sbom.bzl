"""An example of gathering and processing just license information."""

load(
    "@package_metadata//:defs.bzl",
    "PackageAttributeInfo",
    "PackageMetadataInfo",
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


#
# All top level metadata processing rules will generally wrap gahter with their
# Own aspect to walk the tree. This wrapper is usually not much different than
# the example here. The variation is usually only to provide the set of
# providers we want to collect. This allows for organization specific providers
# to be gathered in the same pass as the canonical ones from suppply_chain.
#
def _gather_metadata_info_impl(target, ctx):
    return gather_metadata_info_common(
        target,
        ctx,
        want_providers = [PackageAttributeInfo, PackageMetadataInfo, LicenseKindInfo],
        provider_factory = TransitiveMetadataInfo,
        null_provider_instance = null_transitive_metadata_info,
        filter_func = should_traverse,
    )

gather_metadata_info = aspect(
    doc = """Collects metadata providers into a single TransitiveMetadataInfo provider.""",
    implementation = _gather_metadata_info_impl,
    attr_aspects = ["*"],
    attrs = {
        "_trace": attr.label(default = "@supply_chain_tools//gather_metadata:trace_target"),
    },
    provides = [TransitiveMetadataInfo],
    apply_to_generating_rules = True,
)


def _sbom_impl(ctx):
    # Gather all metadata and make a report from that

    # TODO: Replace this
    # The code below just dumps the collected metadata providers in a somewhat
    # pretty printed way.  In reality, we need to read the files associated with
    # each attribute to get the real data. So this should be a rule to pass
    # all the files to a helper which generates a formated report.
    # That is clearly a job for another day.

    out = []
    if TransitiveMetadataInfo not in ctx.attr.target:
        fail("Missing metadata for %s" % ctx.attr.target)
    t_m_i = ctx.attr.target[TransitiveMetadataInfo]

    report = []
    report.append("Target: %s" % str(ctx.attr.target.label))
    report.append("Gathered target: %s" % str(t_m_i.target))

    command = ["echo"]
    command.append("--target '%s'" % str(t_m_i.target))
    command.append("--output '%s'" % ctx.outputs.out.path)
    for item in t_m_i.metadata.to_list():
       kind = item.kind if hasattr(item, "kind") else "_UNKNOWN_"
       command.append("-kind %s" % kind)
       if hasattr(item, "files"):
           for file in item.files.to_list():
               command.append("-info %s" % file.path)
       # Check for extras
       # This is for debugging during initial development. There should be
       # no extra fields.
       for field in sorted(dir(item)):
           if field in ("files", "kind"):
               continue
           value = getattr(item, field)
           if field == "data":
               report.append("path: %s" % value.path)
           else:
               report.append("%s: %s" % (field, value))

    # TBD: Run the SBOM generator here.
    print("RUN THE SBOM\n  %s\n" % ' '.join(command))

    # This just gives us an output.  Next pass the write will happen in the
    # action we create
    ctx.actions.write(ctx.outputs.out, "\n".join(report) + "\n")
    return [DefaultInfo(files = depset([ctx.outputs.out]))]


"""
  ==== struct(
    metadata = depset([
	struct(
            attributes = <generated file examples/vendor/constant_gen/license_for_emitted_code.package-attribute.json>,
            files = depset([<source file examples/vendor/constant_gen/LICENSE_OF_OUTPUT>, <generated file examples/vendor/constant_gen/license_for_emitted_code.package-attribute.json>])
,
             kind = "build.bazel.attribute.license"),
        struct(
	    files = depset([<source file examples/vendor/acme/ACME_LICENSE>, <generated file examples/vendor/acme/license.package-attribute.json>, <generated file examples/vendor/acme/package_data.package-metadata.json>]),
	    metadata = <generated file examples/vendor/acme/package_data.package-metadata.json>),
        struct(
            attributes = <generated file examples/vendor/libhhgttg/license.package-attribute.json>,
            files = depset([<source file examples/vendor/libhhgttg/LICENSE>,
                         <generated file examples/vendor/libhhgttg/license.package-attribute.json>]),
            kind = "build.bazel.attribute.license"),
        struct(
	    files = depset([<source file LICENSE>, <generated file license.package-attribute.json>, <generated file package_metadata.package-metadata.json>]),
            metadata = <generated file package_metadata.package-metadata.json>)]
      ),
      target = Label("//examples/src:my_violating_server"), traces = [])
   kind: build.bazel.attribute.license, attributes: <generated file examples/vendor/constant_gen/license_for_emitted_code.package-attribute.json>
   kind: <unknown>, metadata: <generated file examples/vendor/acme/package_data.package-metadata.json>
   kind: build.bazel.attribute.license, attributes: <generated file examples/vendor/libhhgttg/license.package-attribute.json>
   kind: <unknown>, metadata: <generated file package_metadata.package-metadata.json>
"""


_sbom = rule(
    implementation = _sbom_impl,
    doc = """Internal tmplementation method for sbom().""",
    attrs = {
        "target": attr.label(
            doc = """Targets to build an SBOM for.""",
            aspects = [gather_metadata_info],
        ),
        "out": attr.output(
            doc = """Output file.""",
            mandatory = True,
        ),
    },
)

def sbom(name, target, out = None, **kwargs):
    """Collects metadata providers for a set of targets and writes a minimal SBOM.

    Args:
      name: The target.
      target: A target to build an SBOM for.
      out: The output file name. Default: <name>.json.
      **kwargs: Other args

    Usage:

      sbom(
          name = "my_app_sbom",
          target = [":my_app"],
          out = "my_app_sbom.json",
      )
    """
    if not out:
        out = name + ".txt"
    _sbom(name = name, target = target, out = out, **kwargs)
