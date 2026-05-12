"""Graph-only metadata serialization.

This module provides functions to serialize TransitiveMetadataInfo providers
to JSON in a graph-only format that avoids duplicating attributes already
present in per-package metadata files. The aggregate JSON contains only the
dependency graph structure (nodes + edges), while attributes remain in
individual metadata files as the single source of truth.
"""

load(":providers.bzl", "TransitiveMetadataInfo")

def _strip_null_repo(label):
    """Removes the null repo name (e.g. @//) from a string.

    The idea is to make str(label) compatible between bazel 5.x and 6.x

    Args:
        label: A Label or string representation of a label

    Returns:
        String with null repo prefix removed
    """
    s = str(label)
    if s.startswith("@//"):
        return s[1:]
    elif s.startswith("@@//"):
        return s[2:]
    return s

def _escape_json_string(s):
    """Escapes a string for JSON output.

    Args:
        s: String to escape

    Returns:
        JSON-safe string
    """
    if not s:
        return ""

    # Replace special characters
    s = s.replace("\\", "\\\\")  # Backslash must be first
    s = s.replace('"', '\\"')  # Double quote
    s = s.replace("\n", "\\n")  # Newline
    s = s.replace("\r", "\\r")  # Carriage return
    s = s.replace("\t", "\\t")  # Tab
    return s

def _get_metadata_file_path(target_with_metadata_info):
    """Extract metadata file path from providers.

    Args:
        target_with_metadata_info: TargetWithMetadataInfo provider

    Returns:
        String path to the .package-metadata.json file, or None if not found
    """
    for provider in target_with_metadata_info.metadata.to_list():
        # PackageMetadataInfo has a 'metadata' field pointing to the JSON file
        if hasattr(provider, "metadata") and hasattr(provider.metadata, "path"):
            return provider.metadata.path
    return None

def _build_node(target_with_metadata_info):
    """Build a graph node for a target.

    Args:
        target_with_metadata_info: TargetWithMetadataInfo provider

    Returns:
        Dictionary representing a node in the graph
    """
    metadata_file = _get_metadata_file_path(target_with_metadata_info)
    return {
        "label": _strip_null_repo(target_with_metadata_info.target),
        "metadata_file": metadata_file if metadata_file else "",
    }

def _build_edges(all_twmi):
    """Build edges from TargetWithMetadataInfo list.

    Args:
        all_twmi: List of TargetWithMetadataInfo providers

    Returns:
        List of edge dictionaries
    """
    edges = []
    for twmi in all_twmi:
        if hasattr(twmi, "direct_deps") and twmi.direct_deps:
            from_label = _strip_null_repo(twmi.target)
            for dep in twmi.direct_deps:
                edges.append({
                    "from": from_label,
                    "to": _strip_null_repo(dep),
                    "type": "depends_on",
                })
    return edges

def _serialize_value_to_json(value):
    """Serialize a value to JSON string.

    Args:
        value: Any value (string, int, bool, list, dict)

    Returns:
        JSON representation
    """
    if value == None:
        return "null"
    elif type(value) == type(""):
        return '"{}"'.format(_escape_json_string(value))
    elif type(value) == type(0):
        return str(value)
    elif type(value) == type(True):
        return "true" if value else "false"
    elif type(value) == type([]):
        if not value:
            return "[]"
        items = [_serialize_value_to_json(item) for item in value]
        return "[\n      " + ",\n      ".join(items) + "\n    ]"
    elif type(value) == type({}):
        return _serialize_dict_to_json(value)
    else:
        # Fallback: convert to string
        return '"{}"'.format(_escape_json_string(str(value)))

def _serialize_dict_to_json(d):
    """Serialize a dict to JSON string.

    Args:
        d: Dictionary

    Returns:
        JSON object string
    """
    if not d:
        return "{}"

    items = []
    for key in sorted(d.keys()):
        value = d[key]
        value_json = _serialize_value_to_json(value)
        items.append('"{key}": {value}'.format(
            key = _escape_json_string(key),
            value = value_json,
        ))

    return "{\n    " + ",\n    ".join(items) + "\n  }"

def metadata_info_to_json(metadata_info):
    """Convert TransitiveMetadataInfo to graph-only JSON.

    This is the main entry point for serialization. It produces JSON with
    only the dependency graph structure (nodes + edges). Attributes are
    NOT duplicated in this output - they remain in per-package metadata files.

    Output format:
    {
      "schema_version": "1.0",
      "root_target": "//some:target",
      "nodes": [
        {
          "label": "//some:target",
          "metadata_file": "path/to/target.package-metadata.json"
        }
      ],
      "edges": [
        {
          "from": "//some:target",
          "to": "//other:dep",
          "type": "depends_on"
        }
      ]
    }

    Args:
        metadata_info: TransitiveMetadataInfo provider

    Returns:
        List containing a single JSON string representation
        (list for backwards compatibility with write_metadata_info)
    """

    # Extract all targets from the transitive depset
    all_twmi = metadata_info.transitive.to_list()

    # Build nodes (sorted by label for deterministic output)
    nodes = [_build_node(twmi) for twmi in sorted(all_twmi, key = lambda x: str(x.target))]

    # Build edges
    edges = _build_edges(all_twmi)

    # Build the output structure
    output = {
        "schema_version": "1.0",
        "root_target": _strip_null_repo(metadata_info.top_level_target) if metadata_info.top_level_target else "",
        "nodes": nodes,
        "edges": edges,
    }

    # Serialize to JSON string
    json_str = _serialize_dict_to_json(output)

    return [json_str]

def write_metadata_info(ctx, deps, json_out):
    """Writes TransitiveMetadataInfo providers for a set of targets as JSON.

    This function writes graph-only JSON that contains dependency structure
    but not attribute duplication. Per-package metadata files remain the
    single source of truth for attributes.

    Usage:
      write_metadata_info must be called from a rule implementation, where the
      rule has run the gather_metadata_info aspect on its deps to
      collect the transitive closure of metadata providers into a
      TransitiveMetadataInfo provider.

      foo = rule(
        implementation = _foo_impl,
        attrs = {
           "deps": attr.label_list(aspects = [gather_metadata_info])
        }
      )

      def _foo_impl(ctx):
        ...
        out = ctx.actions.declare_file("%s_metadata.json" % ctx.label.name)
        write_metadata_info(ctx, ctx.attr.deps, out)

    Args:
      ctx: context of the caller
      deps: a list of deps which should have TransitiveMetadataInfo providers.
            This requires that you have run the gather_metadata_info
            aspect over them
      json_out: output handle to write the JSON info
    """
    metadata_jsons = []
    for dep in deps:
        if TransitiveMetadataInfo in dep:
            metadata_jsons.extend(metadata_info_to_json(dep[TransitiveMetadataInfo]))

    # Wrap multiple roots in an array
    if len(metadata_jsons) == 1:
        content = metadata_jsons[0]
    else:
        content = "[\n" + ",\n".join(metadata_jsons) + "\n]"

    ctx.actions.write(
        output = json_out,
        content = content,
    )
