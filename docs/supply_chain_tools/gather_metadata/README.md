# Metadata Gathering System

The metadata gathering system walks Bazel dependency graphs and collects metadata providers
(licenses, package information, etc.) into a unified format suitable for SBOM generation,
compliance reporting, and supply chain analysis.

## Overview

The system consists of three main components:

1. **Aspect** (`gather_metadata_info`) - Walks the dependency graph collecting metadata
2. **Providers** (`TransitiveMetadataInfo`, `TargetWithMetadataInfo`) - Store collected metadata
3. **Serialization** - Exports graph structure to JSON

## Quick Start

### Collecting Metadata with Aspect

```bash
bazel build //your:target \
  --aspects=@supply_chain_tools//gather_metadata:gather_metadata.bzl%gather_metadata_info_and_write \
  --output_groups=licenses
```

This generates `bazel-bin/your/target_metadata_info.json` with the dependency graph.

### Using in a Rule

```python
load("@supply_chain_tools//gather_metadata:gather_metadata.bzl", "gather_metadata_info")
load("@supply_chain_tools//gather_metadata:serialization.bzl", "write_metadata_info")

my_analysis_rule = rule(
    implementation = _my_analysis_impl,
    attrs = {
        "target": attr.label(aspects = [gather_metadata_info]),
    },
)

def _my_analysis_impl(ctx):
    # Access the provider
    info = ctx.attr.target[TransitiveMetadataInfo]
    
    # Write JSON output
    json_out = ctx.actions.declare_file("%s_metadata.json" % ctx.label.name)
    write_metadata_info(ctx, [ctx.attr.target], json_out)
    
    return [DefaultInfo(files = depset([json_out]))]
```

## Output Format

The serialization produces a **graph-only JSON format** that contains dependency structure
but not attribute data (to avoid duplication with per-package metadata files).

### Aggregate Metadata Graph

```json
{
  "schema_version": "1.0",
  "root_target": "//app:main",
  "nodes": [
    {
      "label": "//app:main",
      "metadata_file": "bazel-bin/app/main.package-metadata.json"
    },
    {
      "label": "//lib:foo",
      "metadata_file": "bazel-bin/lib/foo.package-metadata.json"
    }
  ],
  "edges": [
    {
      "from": "//app:main",
      "to": "//lib:foo",
      "type": "depends_on"
    }
  ]
}
```

**Key points:**
- `nodes` contain target labels and references to metadata files
- `edges` capture dependency relationships (from→to)
- **No attribute duplication** - attributes are in per-package metadata files
- Enables dependency analysis, vulnerability tracking, compliance reporting

### Per-Package Metadata Files

Individual package metadata files (referenced by `metadata_file` in nodes) contain:

```json
{
  "label": "//app:main",
  "purl": "pkg:generic/app@1.0.0",
  "attributes": {
    "package_info": "bazel-bin/app/main.package_info.json",
    "license": "bazel-bin/app/LICENSE.json"
  }
}
```

## Architecture

### Data Flow

```
Bazel Build Graph
    ↓
gather_metadata_info aspect (traverses deps)
    ↓
TargetWithMetadataInfo (per target)
    ↓
TransitiveMetadataInfo (aggregated)
    ↓
metadata_info_to_json (serialization)
    ↓
Graph JSON (nodes + edges)
```

### Providers

**`TargetWithMetadataInfo`**:
- `target`: Label of the target
- `metadata`: Depset of metadata providers attached to this target
- `direct_deps`: List of direct dependency labels (for edges)

**`TransitiveMetadataInfo`**:
- `transitive`: Depset of all `TargetWithMetadataInfo` in the transitive closure
- `top_level_target`: Root of the dependency tree

### Aspect Behavior

The `gather_metadata_info` aspect:
1. Traverses all dependencies via `attr_aspects = ["*"]`
2. Filters attributes based on rule type (see `rule_filters.bzl`)
3. Collects metadata providers (`PackageMetadataInfo`, `LicenseKindInfo`, etc.)
4. Tracks dependency edges for graph generation
5. Returns `TransitiveMetadataInfo` with complete graph

## Use Cases

### SBOM Generation

The dependency graph enables generating SBOMs with relationships:

```python
# Read aggregate
with open("target_metadata_info.json") as f:
    graph = json.load(f)

# Build SBOM with relationships
for edge in graph["edges"]:
    add_dependency_relationship(edge["from"], edge["to"])
```

### Vulnerability Analysis

Find paths from root to vulnerable package:

```python
def find_paths_to_package(graph, target_label):
    """Find all paths from root to a specific package."""
    paths = []
    visited = set()
    
    def dfs(node, path):
        if node in visited:
            return
        visited.add(node)
        path.append(node)
        
        if node == target_label:
            paths.append(list(path))
        
        for edge in graph["edges"]:
            if edge["from"] == node:
                dfs(edge["to"], path)
        
        path.pop()
    
    dfs(graph["root_target"], [])
    return paths
```

### License Compliance

Track license propagation through dependencies:

```python
def check_license_compatibility(graph):
    """Check if any transitive dependencies have incompatible licenses."""
    for node in graph["nodes"]:
        # Read per-package metadata
        with open(node["metadata_file"]) as f:
            metadata = json.load(f)
        
        # Check licenses
        # (Read license files from metadata["attributes"])
```

## Configuration

### Rule Filtering

The aspect skips certain attributes to avoid traversing irrelevant dependencies.
See `rule_filters.bzl` for the complete list.

Common filtered attributes:
- `_*` (private attributes)
- `srcs`, `hdrs` (source files)
- Tool dependencies in specific contexts

### Custom Metadata

To add custom metadata types:

1. Define a provider with metadata fields
2. Attach it to targets via rules
3. The aspect will automatically collect it
4. Access via `TargetWithMetadataInfo.metadata`

## Performance

The system uses Bazel's depset for efficient graph traversal:
- **Memory**: O(n) where n = unique targets with metadata
- **Time**: O(n) traversal, single visit per target
- **Optimization**: Null singleton reused for targets without metadata

## API Reference

See auto-generated documentation:
- [`gather_metadata.bzl`](./gather_metadata.md) - Aspects and functions
- [`serialization.bzl`](../../tools/gather_metadata/serialization.bzl) - JSON serialization
- [`providers.bzl`](../../tools/gather_metadata/providers.bzl) - Provider definitions

## Examples

See the [`examples/`](../../../examples/) directory:
- `examples/sbom/` - Custom SBOM generation
- `examples/sample_reports/` - License compliance reports
