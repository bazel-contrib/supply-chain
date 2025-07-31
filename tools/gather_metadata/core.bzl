"""Rules and macros for collecting package_metadata providers."""

load(":rule_filters.bzl", "rule_to_excluded_attributes")
load(":trace.bzl", "TraceInfo")
load("@package_metadata//:defs.bzl", "PackageAttributeInfo", "PackageMetadataInfo")

DEBUG_LEVEL = 0

def should_traverse(ctx, attr, user_filters = None):
    """Checks if the dependent attribute should be traversed.

    Args:
      ctx: The aspect evaluation context.
      attr: The name of the attribute to be checked.
      user_filters: Additional dictionary of per-rule attribute filters.

    Returns:
      True iff the attribute should be traversed.
    """
    per_rule_filters = [rule_to_excluded_attributes]
    if user_filters:
        per_rule_filters.append(user_filters)

    for filters in per_rule_filters:
        always_ignored = filters.get("*", [])
        if attr in always_ignored:
            return False
        rule_specific_filter = filters.get(ctx.rule.kind, None)
        if rule_specific_filter:
            if (attr in rule_specific_filter
                or "*" in rule_specific_filter
                or ("_*" in rule_specific_filter and attr.startswith("_"))):
                return False
    return True


def _get_transitive_metadata(
        ctx,
        trans_metadata, 
        trans_package_metadata, 
        trans_deps, 
        provider, 
        filter_func = None,
        traces = None):
    """Gather the provider instances of interest from our children

    Args:
        ctx: the ctx
        # TODO
    """    
    attrs = [attr for attr in dir(ctx.rule.attr)]
    for name in attrs:
        if filter_func and not filter_func(ctx, name):
            if DEBUG_LEVEL > 2:
                print("Triming attribute %s of %s" % (name, ctx.rule.kind))
            continue
        if DEBUG_LEVEL > 4:
            print("CHECKING attribute %s of %s" % (name, ctx.rule.kind))

        attr_value = getattr(ctx.rule.attr, name)
        # Make scalers into a lists for convenience.
        if type(attr_value) != type([]):
            attr_value = [attr_value]

        for dep in attr_value:
            # Ignore anything that isn't a target
            if type(dep) != "Target":
                continue

            # Targets can also include things like input files that won't have the
            # aspect, so we additionally check for the aspect rather than assume
            # it's on all targets.  Even some regular targets may be synthetic and
            # not have the aspect. This provides protection against those outlier
            # cases.
            if provider in dep:
                info = dep[provider]
                #XXif info.deps:
                #XX    trans_deps.append(info.deps)
                if hasattr(info, "traces") and getattr(info, "traces"):
                    for trace in info.traces:
                        traces.append("(" + ", ".join([str(ctx.label), ctx.rule.kind, name]) + ") -> " + trace)

                # We only need one or the other of these stanzas.
                # If we use a polymorphic approach to metadata providers, then
                # this works.
                if hasattr(info, "metadata"):
                    if info.metadata:
                        trans_metadata.append(info.metadata)

                # But if we want more precise type safety, we would have a
                # trans_* for each type of metadata. That is not user
                # extensibile.
                if hasattr(info, "package_info"):
                    if info.package_info:
                        trans_package_metadata.append(info.package_info)


def gather_metadata_info_common(
        target,
        ctx,
        want_providers,
        provider_factory = None,
        null_provider_instance = None,
        filter_func = None
    ):
    """Collect package metadata info from myself and my deps.

    Any single target might directly depend on a package metadata, or depend on
    something that transitively depends on a package metadata, or neither.
    This aspect bundles all those into a single provider. At each level, we add
    in new direct metadagta deps found and forward up the transitive information
    collected so far.

    This is a common abstraction for crawling the dependency graph. It is
    parameterized to allow specifying the provider that is populated with
    results. It is configurable to select only a subset of providers. It
    is also configurable to specify which dependency edges should not
    be traced for the purpose of tracing the graph.

    Args:
      target: The target of the aspect.
      ctx: The aspect evaluation context.
      want_providers: a list of providers of interest
      provider_factory: abstracts the provider returned by this aspect
      null_provider_instance: a singleton instance of the empty provider. Reusing a
          a singleton across a large graph can save significant memory.
      filter_func: a function that returns true IFF the dep edge should be ignored

    Returns:
      provider of parameterized type
    """

    # TODO(aiuto): Consider dropping this hack.
    # A hack until https://github.com/bazelbuild/rules_license/issues/89 is
    # fully resolved. If exec is in the bin_dir path, then the current
    # configuration is probably cfg = exec.
    if "-exec-" in ctx.bin_dir.path:
        return null_provider_instance or provider_factory()

    # First we gather my direct metadata providers
    got_providers = []
    package_info = []
    if DEBUG_LEVEL > 1:
        print("==============================================\n %s (%s) \n" % (target.label, ctx.rule.kind))
    if ctx.rule.kind == "build.bazel.attribute.license":
        # Don't try to gather licenses from the license rule itself. We'll just
        # blunder into the text file of the license and pick up the default
        # attribute of the package, which we don't want.
        pass
    else:
        if hasattr(ctx.rule.attr, "package_metadata"):
            package_metadata = ctx.rule.attr.package_metadata
        else:
            package_metadata = []
        for dep in package_metadata:
            if DEBUG_LEVEL > 1:
                print("checking", dep.label)
            for m_p in want_providers:
                if m_p in dep:
                    got_providers.append(dep[m_p])

    if DEBUG_LEVEL > 0 and got_providers:
        print("  GOT: ", target.label, got_providers)

    # Now gather transitive collection of providers from the children
    # this target depends upon.
    trans_metadata = []
    trans_package_metadata = []
    trans_deps = []
    traces = []
    _get_transitive_metadata(
        ctx = ctx,
        trans_metadata = trans_metadata,
        trans_package_metadata = trans_package_metadata,
        trans_deps = trans_deps,
        provider = provider_factory,
        filter_func = filter_func,
        traces = traces,
    )

    # If this is the target, start the sequence of traces.
    if ctx.attr._trace[TraceInfo].trace and ctx.attr._trace[TraceInfo].trace in str(ctx.label):
        traces = [ctx.attr._trace[TraceInfo].trace]

    # Trim the number of traces accumulated since the output can be quite large.
    # A few representative traces are generally sufficient to identify why a dependency
    # is incorrectly incorporated.
    if len(traces) > 10:
        traces = traces[0:10]

    """ THIS MAY NOT MAKE THE PR
    if got_providers:
        # At this point we have a target and a list of direct pointers to
        # metadata. We bundle those together so we can report the exact targets
        # that are of interest. For example, when we have a license applied to
        # a package, we don't want to just report the license - it is more 
        # useful to report the targets in the package that pull in the
        # license.

        # Since a list cannot be stored in a # depset, even inside a provider,
        # the list is concatenated into a string and will be unconcatenated in
        # the output phase.
        direct_license_uses = [LicensedTargetInfo(
            target_under_license = target.label,
            licenses = ",".join([str(x.label) for x in licenses]),
        )]
    else:
        direct_license_uses = None
    """

    # Efficiently merge them:
    # 

    return [provider_factory(
        target = target.label,
        metadata = depset(tuple(got_providers), transitive = trans_metadata),
        # deps = depset(direct = direct_license_uses, transitive = trans_deps),
        traces = traces,
    )]
