"""Rules and macros for collecting package_metadata providers."""

load("@bazel_features//:features.bzl", "bazel_features")
load("@bazel_skylib//rules:common_settings.bzl", "BuildSettingInfo")
load(":providers.bzl", "TargetWithMetadataInfo")
load(":relationships.bzl", "RelationshipInfo", "relationships_for")
load(":rule_filters.bzl", "rule_to_excluded_attributes")
load(":toolchain_involvement.bzl", "ToolchainInvolvementInfo", "toolchain_involvement_for")

def _is_exec_config(ctx):
    """Determines whether the current configuration is an exec configuration."""
    if bazel_features.rules.is_tool_configuration_public and ctx.configuration.is_tool_configuration():
        return True
    elif ctx.bin_dir.path.endswith("-exec/bin"):  # Bazel 9.0.0 or <8.7.0 with --experimental_platform_in_output_dir
        return True
    elif "-exec-" in ctx.bin_dir.path:
        return True
    return False

def should_traverse(ctx, attr, user_filters = None):
    """Checks if the dependent attribute should be traversed.

    Note for the future: We can vastly inmprove the peformance by
    moving this to a Bazel 9 style aspect traversal filter.

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
            if (attr in rule_specific_filter or
                "*" in rule_specific_filter or
                ("_*" in rule_specific_filter and attr.startswith("_"))):
                return False
    return True

def _attr_targets(attr_value):
    """Flattens a single rule attribute value into a list of its targets.

    Rule attributes that carry dependency edges come in a few shapes, and each
    needs to be flattened differently before we can walk the targets it points
    at:

      * ``label`` attributes resolve to a single ``Target``,
      * ``label_list`` attributes resolve to a ``list`` of ``Target``s,
      * ``label_keyed_string_dict`` (and dict-typed attributes in general)
        resolve to a ``dict`` whose keys and/or values may be ``Target``s or
        lists of ``Target``s.

    Any other attribute type (e.g. strings, ints, bools and their list/dict
    variants) carries no targets and flattens to an empty list.

    Args:
      attr_value: the value of a single rule attribute, as returned by
        ``getattr(ctx.rule.attr, name)``.

    Returns:
      A flat list containing every ``Target`` found in ``attr_value``.
    """
    if type(attr_value) == "Target":
        return [attr_value]
    if type(attr_value) == type([]):
        return [item for item in attr_value if type(item) == "Target"]
    if type(attr_value) == type({}):
        # Inspect both keys and values so that any dict-typed attribute (e.g.
        # label_keyed_string_dict, whose keys are targets) is covered.
        targets = []
        for key, value in attr_value.items():
            for entry in (key, value):
                if type(entry) == "Target":
                    targets.append(entry)
                elif type(entry) == type([]):
                    targets.extend([item for item in entry if type(item) == "Target"])
        return targets
    return []

def _include_build_tools(ctx):
    return (
        hasattr(ctx, "attr") and
        hasattr(ctx.attr, "_include_build_tools") and
        ctx.attr._include_build_tools[BuildSettingInfo].value
    )

def _get_transitive_metadata(
        ctx,
        transitive_depsets,
        provider = None,
        null_provider_instance = None,
        filter_func = None,
        direct_deps = None,
        direct_dep_involvements = None,
        direct_dep_relationships = None):
    """Gather the collection provider instances of interest from our children.

    This is a helper to pull up the collected metadata info from children so
    that we can rebundle into the next level efficiently. It revolves around
    a "collection provider" which is the transitive collected data so far.
    While this method is intended to be generic, it is only built and tested
    with TransitiveMetadataInfo.

    Args:
        ctx: the ctx
        transitive_depsets: (output) list of the depsets in the children
        provider: the transitive collection provider.
        null_provider_instance: a singleton instance of the empty provider.
        filter_func: filter to determine to skip.
        direct_deps: (output) list of direct dependency labels for edge tracking
    """

    # A single dependency Target can be reachable through multiple attributes
    # or repeated positions. Add its transitive metadata once, but preserve one
    # edge for each distinct involvement/relationship classification.
    seen_dep_labels = {}
    seen_edges = {}

    attrs = [attr for attr in dir(ctx.rule.attr)]
    for name in attrs:
        if filter_func and not filter_func(ctx, name):
            continue

        for dep in _attr_targets(getattr(ctx.rule.attr, name)):
            # Targets can also include things like input files that won't have
            # the aspect, so we additionally check for the aspect rather than
            # assume it's on all targets. Even some regular targets may be
            # synthetic and not have the aspect. This provides protection
            # against those outlier cases.
            if provider in dep:
                info = dep[provider]
                if info != null_provider_instance:
                    label_key = str(dep.label)
                    if label_key not in seen_dep_labels:
                        seen_dep_labels[label_key] = True
                        transitive_depsets.append(info.transitive)

                    involvement = "build" if info.is_exec_config else "runtime"

                    # Custom aspects wrapping this one may not declare the registry.
                    relationship = "depends_on"
                    if hasattr(ctx.attr, "_relationships"):
                        registry = ctx.attr._relationships[RelationshipInfo]
                        rule_module = ""
                        if hasattr(ctx.rule.attr, "_supply_chain_module"):
                            rule_module = ctx.rule.attr._supply_chain_module
                        relationship = relationships_for(
                            registry,
                            rule_module,
                            ctx.rule.kind,
                            name,
                            attrs,
                        )
                    edge_key = label_key + "\0" + involvement + "\0" + relationship
                    if edge_key not in seen_edges:
                        seen_edges[edge_key] = True
                        if direct_deps != None:
                            direct_deps.append(dep.label)
                            if direct_dep_involvements != None:
                                direct_dep_involvements.append(involvement)
                            if direct_dep_relationships != None:
                                direct_dep_relationships.append(relationship)

    # Toolchain edges are exposed to aspects only on Bazel versions supporting
    # toolchains_aspects = ["*"], and only to aspects declaring the registry
    # attribute (custom aspects wrapping this one may not).
    if (bazel_features.rules.supports_toolchains_aspects_star and
        hasattr(ctx.rule, "toolchains") and
        hasattr(ctx.attr, "_toolchain_involvements")):
        registry = ctx.attr._toolchain_involvements[ToolchainInvolvementInfo]
        for toolchain_type in ctx.rule.toolchains.toolchain_types():
            resolved_toolchain = ctx.rule.toolchains[toolchain_type]

            # Optional toolchain types may not resolve to anything.
            if resolved_toolchain == None or provider not in resolved_toolchain:
                continue
            info = resolved_toolchain[provider]
            if info == null_provider_instance:
                continue

            # `ToolchainAspectsProviders.label` is missing before Bazel 9.
            label = getattr(resolved_toolchain, "label", None) or info.top_level_target
            label_key = str(label)
            if label_key not in seen_dep_labels:
                seen_dep_labels[label_key] = True
                transitive_depsets.append(info.transitive)

            involvement = toolchain_involvement_for(
                registry,
                ctx.rule.kind,
                toolchain_type,
            )
            if involvement == None:
                involvement = "build" if info.is_exec_config else "runtime"
            relationship = "depends_on"
            edge_key = label_key + "\0" + involvement + "\0" + relationship
            if edge_key not in seen_edges:
                seen_edges[edge_key] = True
                if direct_deps != None:
                    direct_deps.append(label)
                    if direct_dep_involvements != None:
                        direct_dep_involvements.append(involvement)
                    if direct_dep_relationships != None:
                        direct_dep_relationships.append(relationship)

def gather_metadata_info_common(
        target,
        ctx,
        want_providers = None,
        provider_factory = None,
        null_provider_instance = None,
        filter_func = None,
        enable_intermediate_pruning = True):
    """Collect package metadata info from myself and my deps.

    Any single target might directly depend on a package metadata, or depend on
    something that transitively depends on a package metadata, or neither.
    This aspect bundles all those into a single provider. At each level, we add
    in new direct metadata deps found and forward up the transitive information
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
      enable_intermediate_pruning: if true, targets without direct metadata are
          collapsed into the compact transitive representation. Package graphs
          need these intermediate targets retained, so they should set this false.

    Returns:
      provider of parameterized type
    """

    # TODO(aiuto): Consider dropping this hack.
    # A hack until https://github.com/bazelbuild/rules_license/issues/89 is
    # fully resolved.
    if _is_exec_config(ctx) and not _include_build_tools(ctx):
        return [null_provider_instance or provider_factory()]

    # First we gather my direct metadata providers.
    # This captures the pairs if
    got_providers = []
    package_info = []
    if hasattr(ctx.rule.attr, "kind") and ctx.rule.attr.kind == "build.bazel.attribute.license":
        # Don't try to gather licenses from the license rule itself. We'll just
        # blunder into the text file of the license and pick up the default
        # attribute of the package, which we don't want.
        pass
    else:
        if hasattr(ctx.rule.attr, "package_metadata"):
            package_metadata = ctx.rule.attr.package_metadata
        elif hasattr(ctx.rule.attr, "applicable_licenses"):
            package_metadata = ctx.rule.attr.applicable_licenses
        else:
            package_metadata = []
        for metadata_dependency in package_metadata:
            for wanted_provider in want_providers:
                if wanted_provider in metadata_dependency:
                    got_providers.append(metadata_dependency[wanted_provider])

    # Now gather transitive collection of providers from the children this
    # target depends upon.
    transitive_depsets = []
    direct_deps = []
    direct_dep_involvements = []
    direct_dep_relationships = []
    _get_transitive_metadata(
        ctx = ctx,
        transitive_depsets = transitive_depsets,
        provider = provider_factory,
        null_provider_instance = null_provider_instance,
        filter_func = filter_func,
        direct_deps = direct_deps,
        direct_dep_involvements = direct_dep_involvements,
        direct_dep_relationships = direct_dep_relationships,
    )

    # State so far:
    # got_providers: list (maybe empty) of metadata providers we directly have
    # transitive_depsets: the list of the collection providers from our children.

    # Efficiently merge them.

    # We can do some tricks to avoid allocating a lot of memory
    # in big graphs. For the most part, metadata attachments are near the
    # leaves, and sparse higher up.
    # 1. If there is no direct metadata (got_providers is None) and there is
    #    no transitive metadata, return the null instance.  This is typical
    #    for home grown code that only depends on our own code.
    # 2. If got_providers is None, and there transitive info.
    #    If the length of the list is one, just pass up the first element.
    #    This is common through the whole middle of a build graph.
    # 3. If the above fail, construct a new one.

    if not got_providers and not transitive_depsets:
        return [null_provider_instance or provider_factory()]

    # Package graphs need intermediate targets even when they have no metadata,
    # so they disable the fast path. Others can use the compact representation.
    if not got_providers and enable_intermediate_pruning:
        """
        TODO: If there is only one, pass up the entire provider, not the extracted transitive
        if len(transitive_depsets) == 1 and transitive_depsets[0]:
            # Often, there is only one thing we are passing up. There is no
            # reason to allocate another collection provider around that.
            return transitive_depsets[0]
         """
        return [provider_factory(
            transitive = depset(transitive = transitive_depsets),
            top_level_target = target.label,
            is_exec_config = _is_exec_config(ctx),
        )]

    # Create a TWMI linking this target to the applicable metadata
    me = TargetWithMetadataInfo(
        target = target.label,
        metadata = depset(got_providers),
        direct_deps = tuple(direct_deps),  # Convert to tuple for immutability (required for depset)
        direct_dep_involvements = tuple(direct_dep_involvements),
        direct_dep_relationships = tuple(direct_dep_relationships),
    )
    if not transitive_depsets:
        return [provider_factory(
            transitive = depset(direct = [me]),
            top_level_target = target.label,
            is_exec_config = _is_exec_config(ctx),
        )]
    return [provider_factory(
        transitive = depset(direct = [me], transitive = transitive_depsets),
        top_level_target = target.label,
        is_exec_config = _is_exec_config(ctx),
    )]
