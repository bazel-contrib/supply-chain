# Vendored SBOM schemas

These files are vendored verbatim (byte-for-byte, no modifications) from
upstream so that schema validation of generated SBOMs does not require
network access at build/test time. Each file still declares its original
`$id`; the validator (`//lib/supplychain-go/cmd/schemavalidate`) registers
every file it's given under that `$id`, so cross-file `$ref`s resolve
locally rather than hitting the network.

To refresh, re-fetch the same paths at a newer tag and update this file's
pinned commit/tag.

## CycloneDX (`cyclonedx/`)

Source: https://github.com/CycloneDX/specification
Pinned tag: `1.6.1` (commit `8a27bfd1be5be0dcb2c208a34d2f4fa0b6d75bd7`)

- `bom-1.6.schema.json` — `schema/bom-1.6.schema.json`, the root BOM schema.
  Matches the spec version `cyclonedx-go` (our dependency) defaults to for
  `cdx.NewBOM()` (`SpecVersion1_6`).
- `spdx.schema.json` — `schema/spdx.schema.json`, referenced by
  `bom-1.6.schema.json` for the SPDX license ID enum.
- `jsf-0.82.schema.json` — `schema/jsf-0.82.schema.json`, referenced by
  `bom-1.6.schema.json` for the `signature` definition (JSON Signature
  Format).

## SPDX (`spdx/`)

Source: https://github.com/spdx/spdx-spec
Pinned tag: `v2.3` (commit `aadf3b0b8dbbabdb4d880b0fc714255fea436ff7`)

- `spdx-schema.json` — `schemas/spdx-schema.json`, self-contained (no
  external `$ref`s).

Note: `cmd/spdx` generates **SPDX 2.3** documents
(`spdx.Document{SPDXVersion: "SPDX-2.3", ...}`), so that's the schema vendored
here. The `develop` branch of `spdx/spdx-spec` linked when this was requested
holds the **SPDX 3.0** schemas, which use an entirely different (JSON-LD
based) data model and would not validate our 2.3 output. If `cmd/spdx` is
ever upgraded to emit SPDX 3.0, the vendored schema (and validator wiring)
needs to move to the 3.0 schema set, which is split across multiple files
under `model/*/spdx-schema.json` rather than the single-file 2.3 layout.
