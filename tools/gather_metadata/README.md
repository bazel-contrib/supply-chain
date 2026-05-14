# Metadata Gathering System

Bazel rules and aspects to walk dependency trees and gather metadata providers (licenses, package information, etc.).

This set of rules produces a graph of nodes (targets) and their edges starting from a root target. This can be automatically serialized to JSON and looks like the example below:

```json
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
```