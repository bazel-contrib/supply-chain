<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Declares rule `patch`.

<a id="PatchInfo"></a>

## PatchInfo

<pre>
load("@package_metadata//attributes:patch.bzl", "PatchInfo")

PatchInfo(<a href="#PatchInfo-value">value</a>)
</pre>

Provider carrying the fields of a single `patch` target, consumed by `patch_set`.

**FIELDS**

| Name  | Description |
| :------------- | :------------- |
| <a id="PatchInfo-value"></a>value |  A dict holding the JSON-serializable fields of the patch.    |


<a id="patch"></a>

## patch

<pre>
load("@package_metadata//attributes:patch.bzl", "patch")

patch(*, <a href="#patch-name">name</a>, <a href="#patch-description">description</a>, <a href="#patch-integrity">integrity</a>, <a href="#patch-type">type</a>, <a href="#patch-url">url</a>, <a href="#patch-file">file</a>, <a href="#patch-resolves">resolves</a>, <a href="#patch-tags">tags</a>, <a href="#patch-visibility">visibility</a>)
</pre>

Rule for declaring a single patch applied to a package.

Bazel attributes cannot be structs, so a patch is declared as its own
target and grouped with `patch_set`. This target can also be passed on its
own to the `attributes` of a `package_metadata` target.

Values are passed through as-is (apart from `type`): tools consuming the
attribute are responsible for mapping them to the format they produce.
Entries are assumed to be UTF-8 encoded.

Usage:

```starlark
load("@package_metadata//attributes:patch.bzl", "patch")

patch(
    name = "fix_overflow",
    description = "Fixes a buffer overflow in the parser.",
    integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
    resolves = ["CVE-2024-12345"],
    type = "backport",
    url = "https://example.com/patches/fix_overflow.patch",
)

patch(
    name = "local_tweak",
    description = "Local tweak.",
    file = "patches/local_tweak.patch",
    integrity = ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
)
```

Format: the `PackageAttributeInfo` of this rule has `kind`
`build.bazel.attribute.patch`, and its `attributes` file holds a JSON object
written without insignificant whitespace, with keys sorted. `type`, `url`, `file`
and `resolves` are omitted when unset:

```json
{
    "description": "Fixes a buffer overflow in the parser.",
    "integrity": ["sha256-EbbT7lVO7feSmZBamPm5oE5JghC1nxUJTJFskdFQ51Q="],
    "resolves": ["CVE-2024-12345"],
    "type": "backport",
    "url": "https://example.com/patches/fix_overflow.patch"
}
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="patch-name"></a>name |  A unique name for this target.   |  none |
| <a id="patch-description"></a>description |  Required. Human-readable description of what the patch does.   |  none |
| <a id="patch-integrity"></a>integrity |  Required. Checksums of the patch file as SRI strings (`<algorithm>-<base64 digest>`), e.g. `["sha256-<base64 digest>"]`. Must contain at least one entry.   |  none |
| <a id="patch-type"></a>type |  Optional. The nature of the patch, one of `unofficial`, `monkey`, `backport` or `cherry-pick` (CycloneDX `pedigree.patches[].type`).   |  `None` |
| <a id="patch-url"></a>url |  Where the patch can be retrieved from. Exactly one of `url` or `file` must be set.   |  `None` |
| <a id="patch-file"></a>file |  A patch file in the workspace. Exactly one of `url` or `file` must be set. The JSON holds the file's `short_path`, and the file is added to the `files` of the `PackageAttributeInfo`.   |  `None` |
| <a id="patch-resolves"></a>resolves |  Optional. Identifiers (e.g., CVE ids, issue URLs) of the issues the patch resolves.   |  `None` |
| <a id="patch-tags"></a>tags |  A list of arbitrary tags to apply to this target.   |  `None` |
| <a id="patch-visibility"></a>visibility |  The visibility of this target.   |  `None` |


