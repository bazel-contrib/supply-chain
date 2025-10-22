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
    "TransitiveMetadataInfo",
    "null_transitive_metadata_info",
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

def _handle_provider(metadata_provider, command, inputs, report):
    """Handle an individual metadata provider.

    Args:
        metadata_provider: A provider instance
        command: (in/out) list of command line args we are building
        inputs: (in/out) list of files needed for that command line
        report: (in/out) list of things we want to say to the user.
                This is for illustrating how to use these rules, and
                is not needed for the SBOM.
    """

    # We are presuming having metadata means you are a PackageMetadataInfo.
    if hasattr(metadata_provider, "metadata"):
        command.append("-metadata %s" % metadata_provider.metadata.path)
        inputs.extend(metadata_provider.files.to_list())
        if hasattr(metadata_provider, "purl"):
            command.append("-purl %s" % metadata_provider.purl)
            report.append("purl %s" % metadata_provider.purl)
        return

    # If you are gathering your own custom types, having a kind field can
    # be used to distingish them if they need different post processing.
    kind = metadata_provider.kind if hasattr(metadata_provider, "kind") else "_UNKNOWN_"
    print("##-- %s" % kind)
    print(metadata_provider)
    if kind:
        # but maybe the kind is in the info file.
        command.append("-kind %s" % kind)
        if hasattr(metadata_provider, "attributes"):
            command.append("-attributes %s" % metadata_provider.attributes.path)
            report.append("  Attribute data: %s" % metadata_provider.attributes.short_path)
            if hasattr(metadata_provider, "files"):
                inputs.extend(metadata_provider.files.to_list())
                for f in metadata_provider.files.to_list():
                    report.append("    file: %s" % f.short_path)

def _handle_trans_collector(t_m_i, command, inputs, report):
    """Process a TransitiveMetadataInfo.

    Args:
        t_m_i: A provider instance
        command: (in/out) list of command line args we are building
        inputs: (in/out) list of files needed for that command line
        report: (in/out) list of things we want to say to the user.
                This is for illustrating how to use these rules, and
                is not needed for the SBOM.
    """
    if hasattr(t_m_i, "directs"):
        print("HAS DIRECTS")
        print(t_m_i.directs.to_list())
        for direct in t_m_i.directs.to_list():
            _handle_provider(direct, command, inputs, report)
    if hasattr(t_m_i, "trans"):
        print("HAS TRANS")
        for trans in t_m_i.trans.to_list():
            if hasattr(trans, "directs"):
                print("Inner DIRECTS")
                for direct in trans.directs.to_list():
                    _handle_provider(direct, command, inputs, report)
            if hasattr(trans, "trans"):
                print("=======Trans in trans")
                print(trans.trans)
                print(">>>>>>")
                # _handle_trans_collector(trans, command, inputs, report)

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
    print(t_m_i)

    inputs = []
    report = []
    command = ["echo"]
    command.append("--output '%s'" % ctx.outputs.out.path)

    report.append("Target: %s" % str(ctx.attr.target.label))
    if hasattr(t_m_i, "target"):
        report.append("Gathered target: %s" % str(t_m_i.target))
        command.append("--target '%s'" % str(t_m_i.target))

    if hasattr(t_m_i, "directs"):
        print("TOP HAS DIRECTS")
        for direct in t_m_i.directs.to_list():
            _handle_provider(direct, command, inputs, report)
    if hasattr(t_m_i, "trans"):
        print("TOP HAS TRANS")
        for trans in t_m_i.trans.to_list():
            _handle_trans_collector(trans, command, inputs, report)

    # TBD: Run the SBOM generator here.
    print("RUN THE SBOM\n  %s\n" % " ".join(command))
    print("Report: \n   %s\n" % "\n   ".join(report))

    # This just gives us an output.  Next pass the write will happen in the
    # action we create
    ctx.actions.write(ctx.outputs.out, "\n".join(report) + "\n")
    return [DefaultInfo(files = depset([ctx.outputs.out]))]

_sbom = rule(
    implementation = _sbom_impl,
    doc = """Internal tmplementation method for sbom().""",
    attrs = {
        "out": attr.output(
            doc = """Output file.""",
            mandatory = True,
        ),
        "target": attr.label(
            doc = """Targets to build an SBOM for.""",
            aspects = [gather_metadata_info],
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
