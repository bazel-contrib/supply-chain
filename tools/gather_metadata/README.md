# Metadata Gathering System

Bazel rules and aspects to walk dependency trees and gather metadata providers
(licenses, package information, etc.).

## Documentation

See comprehensive documentation at [docs/supply_chain_tools/gather_metadata/](../../docs/supply_chain_tools/gather_metadata/README.md)

## Quick Start

```bash
bazel build //your:target \
  --aspects=@supply_chain_tools//gather_metadata:gather_metadata.bzl%gather_metadata_info_and_write \
  --output_groups=licenses
```

## Files

- `gather_metadata.bzl` - Main aspect definitions
- `serialization.bzl` - JSON serialization (graph-only format)
- `core.bzl` - Core aspect traversal logic
- `providers.bzl` - Provider definitions
- `rule_filters.bzl` - Attribute filtering rules
- `serialization_test.bzl` - Unit tests
- `integration_test.bzl` - Integration tests
