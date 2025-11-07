"""Gather licenses used by bazel targets."""

load(
    "@package_metadata//:defs.bzl",
    "PackageAttributeInfo",
    "PackageMetadataInfo",
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

DEBUG_LEVEL = 0

def update_attribute_to_consumers(attribute_to_consumers, file, target):
    if file.path not in attribute_to_consumers:
        attribute_to_consumers[file.path] = []
    attribute_to_consumers[file.path].append(str(target))

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
        want_providers = [PackageAttributeInfo, PackageMetadataInfo],
        provider_factory = TransitiveMetadataInfo,
        null_provider_instance = null_transitive_metadata_info,
        filter_func = should_traverse,
    )

gather_metadata_info = aspect(
    doc = """Collects metadata providers into a single TransitiveMetadataInfo provider.""",
    implementation = _gather_metadata_info_impl,
    attr_aspects = ["*"],
    provides = [TransitiveMetadataInfo],
    apply_to_generating_rules = True,
)

def _handle_attribute_provider(metadata_provider, target = None, command = None, inputs = None, report = None, attribute_to_consumers = None):
    """Handle an individual metadata provider.

    Args:
        metadata_provider: A provider instance
        target: target to which this attribute applies
        command: (in/out) list of command line args we are building
        inputs: (in/out) list of files needed for that command line
        report: (in/out) list of things we want to say to the user.
                This is for illustrating how to use these rules, and
                is not needed for the SBOM.
        attribute_to_consumers: Map of attribute providers back to the packages that use them.
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
    if DEBUG_LEVEL > 1:
        # buildifier: disable=print
        print("##-- %s" % kind)

        # buildifier: disable=print
        print(metadata_provider)
    if kind:
        # but maybe the kind is in the info file.
        command.append("-kind %s" % kind)

        if hasattr(metadata_provider, "attributes"):
            update_attribute_to_consumers(attribute_to_consumers, metadata_provider.attributes, target)
            command.append("-attributes %s" % metadata_provider.attributes.path)
            report.append("  Attribute data: %s" % metadata_provider.attributes.short_path)
            if hasattr(metadata_provider, "files"):
                inputs.extend(metadata_provider.files.to_list())
                for f in metadata_provider.files.to_list():
                    report.append("    file: %s" % f.path)

        # Check for extras.
        # This is for debugging during initial development. There should be
        # no extra fields.
        for field in sorted(dir(metadata_provider)):
            if field in ("attributes", "files", "kind"):
                continue
            value = getattr(metadata_provider, field)
            report.append("%s: %s" % (field, value))

def _handle_trans_collector(t_m_i, command, inputs, report, attribute_to_consumers):
    """Process a TransitiveMetadataInfo.

    Args:
        t_m_i: A provider instance
        command: (in/out) list of command line args we are building
        inputs: (in/out) list of files needed for that command line
        report: (in/out) list of things we want to say to the user.
                This is for illustrating how to use these rules, and
                is not needed for the SBOM.
        attribute_to_consumers: Map of attribute providers back to the
                 packages that use them.
    """
    if hasattr(t_m_i, "metadata"):
        report.append("Target %s" % t_m_i.target)
        command.append("-target %s" % t_m_i.target)
        for metadata in t_m_i.metadata.to_list():
            _handle_attribute_provider(
                metadata,
                target = t_m_i.target,
                command = command,
                inputs = inputs,
                report = report,
                attribute_to_consumers = attribute_to_consumers,
            )
        if hasattr(t_m_i, "trans"):
            fail("TransititiveMetadataInfo contains both metadata and trans." + str(t_m_i))

def _licenses_used_impl(ctx):
    # Gather all metadata and make a report from that

    # TODO: Replace this
    # The code below just dumps the collected metadata providers in a somewhat
    # pretty printed way.  In reality, we need to read the files associated with
    # each attribute to get the real data. So this should be a rule to pass
    # all the files to a helper which generates a formated report.
    # That is clearly a job for another day.

    if TransitiveMetadataInfo not in ctx.attr.target:
        fail("Missing metadata for %s" % ctx.attr.target)
    t_m_i = ctx.attr.target[TransitiveMetadataInfo]
    if DEBUG_LEVEL > 0:
        # buildifier: disable=print
        print(t_m_i)

    inputs = []
    report = []
    attribute_to_consumers = {}

    command = ["echo"]
    command.append("--output '%s'" % ctx.outputs.out.path)

    report.append("Top label: %s" % str(ctx.attr.target.label))
    if hasattr(t_m_i, "target"):
        report.append("Target: %s" % str(t_m_i.target))
        command.append("--target '%s'" % str(t_m_i.target))

    # It is possible for the top level target to have metadata, but rare.
    if hasattr(t_m_i, "metadata"):
        if DEBUG_LEVEL > 0:
            # buildifier: disable=print
            print("TOP HAS DIRECTS")
        for direct in t_m_i.metadata.to_list():
            _handle_attribute_provider(
                metadata = direct,
                target = t_m_i.target,
                command = command,
                inputs = inputs,
                report = report,
                attribute_to_consumers = attribute_to_consumers,
            )

    if hasattr(t_m_i, "trans"):
        for trans in t_m_i.trans.to_list():
            _handle_trans_collector(trans, command, inputs, report, attribute_to_consumers)
    if DEBUG_LEVEL > 1:
        # buildifier: disable=print
        print(json.encode_indent(attribute_to_consumers))

    # For the time being, print a report of what we have. It's not the final
    # output. It just helps see what we have.
    # buildifier: disable=print
    print("Report: \n   %s\n" % "\n   ".join(report))

    # TBD: Run the generator here.
    # buildifier: disable=print
    print("Run the report maker\n  %s\n" % " ".join(command))

    # This just gives us an output.  Next pass the write will happen in the
    # action we create
    ctx.actions.write(ctx.outputs.out, "\n".join(report) + "\n")
    return [DefaultInfo(files = depset([ctx.outputs.out]))]

licenses_used = rule(
    implementation = _licenses_used_impl,
    doc = """Create a list of the licenses used by a target.""",
    attrs = {
        "target": attr.label(
            doc = """Targets to gather licenses for.""",
            aspects = [gather_metadata_info],
        ),
        "out": attr.output(
            doc = """Output file.""",
            mandatory = True,
        ),
    },
)
