# Tutorial: Adding `package_metadata` as a rule author

This tutorial is for authors of rule sets — in particular rule sets that fetch
third-party packages (e.g., from Maven, npm, PyPI, crates.io, or Go module
proxies) into a Bazel build. It shows how to annotate the repositories your
rules create with `package_metadata`, so that consumers of your rule set can
produce SBOMs and run license checks for everything they pull in through it.

If you only want to annotate your own module, see
[As module author](../metadata/README.md#as-module-author) — this tutorial builds on it.

  - [Background](#background)
  - [Step 1: Depend on `@package_metadata`](#step-1-depend-on-package_metadata)
  - [Step 2: Build the PURL with `purl.builder`](#step-2-build-the-purl-with-purlbuilder)
  - [Step 3: Annotate repositories created by repository rules](#step-3-annotate-repositories-created-by-repository-rules)
  - [Step 4: Annotate `BUILD` file overlays](#step-4-annotate-build-file-overlays)
  - [Step 5 (optional): Write a rule that provides `PackageMetadataInfo`](#step-5-optional-write-a-rule-that-provides-packagemetadatainfo)
  - [Step 6: Verify the metadata](#step-6-verify-the-metadata)
  - [Checklist](#checklist)


## Background

A `package_metadata` target identifies one package by its
[PURL](https://github.com/package-url/purl-spec) and carries `attributes`
about it (e.g., its [license](../metadata/licenses/rules/license.md#license)). Targets are
associated with it in one of three ways:

  - repo-wide, with `repo(default_package_metadata = [...])` in `REPO.bazel`,
  - package-wide, with `package(default_package_metadata = [...])`, or
  - per target, with the `package_metadata` attribute that every rule has.

Tools such as [`@supply_chain_tools`](../../tools) walk the dependency graph of
a target and collect the metadata of every target they visit. Your rules do not
need to do anything to be visited: tagging the targets is enough.

For a rule set that fetches packages, the pattern is therefore:

  1. compute the PURL of each fetched package from the information your rule
     set already has (coordinates, version, registry URL, ...),
  2. generate a `package_metadata` target in the fetched repository, and
  3. point `default_package_metadata` in the repository's `REPO.bazel` at it.


## Step 1: Depend on `@package_metadata`

Add a regular (non-dev) dependency to your `MODULE.bazel`:

```starlark
bazel_dep(name = "package_metadata", version = "<check releases>")
```

`@package_metadata` has no non-dev dependencies of its own, so this does not
add anything else to your consumers' dependency graphs.


## Step 2: Build the PURL with `purl.builder`

The PURL is the identity of a package. SBOM tools and vulnerability scanners
match on it, so it has to be canonical. Build it with
[`purl.builder`](../metadata/purl/purl.md#purlbuilder) instead of formatting a string
yourself:

```starlark
load("@package_metadata//purl:purl.bzl", "purl")

guava = (purl.builder()
    .type("maven")
    .namespace("com.google.guava")
    .name("guava")
    .version("33.4.0-jre")
    .build())
# pkg:maven/com.google.guava/guava@33.4.0-jre
```

The builder takes care of the details that hand-written format strings tend to
get wrong:

  - **Percent-encoding.** Components and qualifier values are encoded as the
    spec requires, e.g., the scope of npm packages or URLs in qualifiers:

    ```starlark
    purl.builder().type("npm").namespace("@babel").name("core").version("7.26.0").build()
    # pkg:npm/%40babel/core@7.26.0
    ```

  - **Type-specific normalization.** For example, PyPI names are lowercased and
    `_` is replaced by `-`; GitHub namespaces and names are lowercased. Pass the
    names as your package manager reports them and let the builder normalize
    them:

    ```starlark
    purl.builder().type("pypi").name("Django_REST").version("3.15").build()
    # pkg:pypi/django-rest@3.15
    ```

  - **Qualifier ordering.** Qualifiers are sorted, so the same package always
    gets the same PURL no matter the order you add them in.

  - **Validation.** `build()` fails if a required component is missing or a
    component is invalid for the type. In a repository rule, this surfaces as
    a fetch error pointing at your rule set rather than as a malformed SBOM
    much later.

Guidelines for choosing the components:

  - **`type`**: use the type of the ecosystem the package is fetched from, from
    the [list of PURL types](https://github.com/package-url/purl-spec/blob/main/purl-types-index.json)
    (e.g., `maven`, `npm`, `pypi`, `cargo`, `golang`, `github`). Bazel modules
    have their own helper, [`purl.bazel`](../metadata/purl/purl.md#purlbazel). If no type
    fits, use `generic` and add a `download_url` or `vcs_url` qualifier so the
    package can be identified unambiguously:

    ```starlark
    (purl.builder()
        .type("generic")
        .name("openssl")
        .version("3.4.0")
        .add_qualifier("download_url", "https://github.com/openssl/openssl/releases/download/openssl-3.4.0/openssl-3.4.0.tar.gz")
        .build())
    ```

  - **`version`**: the exact, resolved version that was fetched, never a range
    or constraint.

  - **`repository_url`**: add this qualifier only when the package was fetched
    from somewhere other than the type's default registry (e.g., a private
    Maven repository instead of Maven Central). This mirrors what `purl.bazel`
    does for registries other than the BCR.

The builder is plain Starlark: it works in `BUILD` files, in rule
implementations, in repository rules, and in module extensions.

Do not call `disable_checks()` unless you have a specific reason to; it can
produce PURLs that parsers reject.


## Step 3: Annotate repositories created by repository rules

This is the main integration point for rule sets that fetch packages. The
example below is a minimal repository rule that downloads a Maven artifact. The
supply-chain-specific parts are:

  - the PURL is computed with `purl.builder` in the repository rule,
  - `REPO.bazel` sets `default_package_metadata`, so every target in the
    repository — including those your rule set or users add later — is
    associated with the package, and
  - the generated `BUILD` file loads the `@package_metadata` rules through
    stringified `Label`s.

```starlark
# acme/private/maven_artifact.bzl
"""Declares repository rule `maven_artifact`."""

load("@package_metadata//purl:purl.bzl", "purl")

_MAVEN_CENTRAL = "https://repo.maven.apache.org/maven2"

# Stringified `Label`s resolve to canonical labels, so the generated `BUILD`
# file can load them no matter which repos are visible to the generated repo.
_LICENSE_BZL = str(Label("@package_metadata//licenses/rules:license.bzl"))
_PACKAGE_METADATA_BZL = str(Label("@package_metadata//rules:package_metadata.bzl"))

def _spdx_license_kind(identifier):
    return str(Label("@package_metadata//licenses/spdx:{}".format(identifier)))

def _maven_purl(*, group, artifact, version, repository):
    builder = (purl.builder()
        .type("maven")
        .namespace(group)
        .name(artifact)
        .version(version))

    # Maven Central is the default repository of the `maven` PURL type, so only
    # other repositories are recorded.
    if repository != _MAVEN_CENTRAL:
        builder.add_qualifier("repository_url", repository)

    return builder.build()

def _maven_artifact_impl(rctx):
    group, artifact, version = rctx.attr.coordinates.split(":")

    rctx.download(
        url = "{repository}/{group_path}/{artifact}/{version}/{artifact}-{version}.jar".format(
            repository = rctx.attr.repository,
            group_path = group.replace(".", "/"),
            artifact = artifact,
            version = version,
        ),
        output = "{}.jar".format(artifact),
        sha256 = rctx.attr.sha256,
    )

    # Associate every target in this repo with `//:package_metadata`.
    rctx.file("REPO.bazel", content = """\
repo(default_package_metadata = ["//:package_metadata"])
""")

    build = [
        "load({}, \"package_metadata\")".format(repr(_PACKAGE_METADATA_BZL)),
    ]
    attributes = []
    if rctx.attr.license:
        build.append("load({}, \"license\")".format(repr(_LICENSE_BZL)))
        attributes.append(":license")

    build.append("""
package_metadata(
    name = "package_metadata",
    attributes = {attributes},
    purl = {purl},
    visibility = ["//visibility:public"],
)
""".format(
        attributes = repr(attributes),
        purl = repr(_maven_purl(
            group = group,
            artifact = artifact,
            version = version,
            repository = rctx.attr.repository,
        )),
    ))

    if rctx.attr.license:
        build.append("""
license(
    name = "license",
    kind = {kind},
)
""".format(
            kind = repr(_spdx_license_kind(rctx.attr.license)),
        ))

    build.append("""
filegroup(
    name = "jar",
    srcs = [{jar}],
    visibility = ["//visibility:public"],
)
""".format(
        jar = repr("{}.jar".format(artifact)),
    ))

    rctx.file("BUILD.bazel", content = "\n".join(build))

maven_artifact = repository_rule(
    implementation = _maven_artifact_impl,
    attrs = {
        "coordinates": attr.string(
            mandatory = True,
            doc = "The Maven coordinates (`group:artifact:version`) of the artifact.",
        ),
        "license": attr.string(
            doc = "The SPDX identifier of the license of the artifact, if known.",
        ),
        "repository": attr.string(
            default = _MAVEN_CENTRAL,
            doc = "The URL of the Maven repository to download the artifact from.",
        ),
        "sha256": attr.string(),
    },
)
```

For `coordinates = "com.google.guava:guava:33.4.0-jre"` and
`license = "Apache-2.0"`, the generated `BUILD.bazel` is:

```starlark
load("@@package_metadata+//rules:package_metadata.bzl", "package_metadata")
load("@@package_metadata+//licenses/rules:license.bzl", "license")

package_metadata(
    name = "package_metadata",
    attributes = [":license"],
    purl = "pkg:maven/com.google.guava/guava@33.4.0-jre",
    visibility = ["//visibility:public"],
)

license(
    name = "license",
    kind = "@@package_metadata+//licenses/spdx:Apache-2.0",
)

filegroup(
    name = "jar",
    srcs = ["guava.jar"],
    visibility = ["//visibility:public"],
)
```

Things to keep in mind:

  - **Compute the PURL in Starlark, emit the result.** Calling `purl.builder`
    in the repository rule (or in the module extension that calls it) validates
    the PURL at fetch time and keeps the generated `BUILD` file free of logic.
    Always pass the result through `repr()` when writing it into a `BUILD`
    file.

  - **Use `default_package_metadata` in `REPO.bazel`.** It covers every target
    in the repository with a single line. If one repository contains several
    packages (e.g., a lock-file-based "hub" repository), generate one
    `package_metadata` target per package and set
    `package(default_package_metadata = ...)` in each package, or the
    `package_metadata` attribute on each target, instead.

  - **Licenses.** Many package managers report an SPDX license identifier (in a
    POM, `package.json`, `Cargo.toml`, ...). Map it to the predeclared
    `license_kind` targets in `@package_metadata//licenses/spdx`. Only emit a
    `license` when the identifier is a single SPDX license ID — that package
    has a target for every ID on the SPDX license list, and analysis fails for
    anything else. SPDX expressions such as `MIT OR Apache-2.0` are not
    modelled yet; leave the `license` out rather than guessing. If the fetched
    archive contains the license text, pass it as `text`:

    ```starlark
    if rctx.path("LICENSE").exists:
        # ... `text = "LICENSE",` in the generated `license` target.
    ```

  - **Load `@package_metadata` through `str(Label(...))`.** The stringified
    label is canonical (e.g., `@@package_metadata+//...`), so the generated
    repository does not depend on which apparent repository names it can see.
    This also works for repositories created from `WORKSPACE`.

  - **Consumers can override what you declare.** If your metadata is missing or
    wrong for a package, consumers can replace it with
    [`package_metadata_override`](../metadata-extensions/rules/package_metadata_override.md#package_metadata_override)
    without patching your rule set. That is a fallback, not a substitute:
    declaring correct metadata in your rule set fixes it for everyone.


## Step 4: Annotate `BUILD` file overlays

Rule sets often ship hand-written `BUILD` files for third-party archives (e.g.,
`http_archive(build_file = ...)`). Annotate those like any other package, and
use `purl.builder` directly in the `BUILD` file:

```starlark
# third_party/zlib.BUILD.bazel
load("@package_metadata//licenses/rules:license.bzl", "license")
load("@package_metadata//purl:purl.bzl", "purl")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package(default_package_metadata = [":package_metadata"])

package_metadata(
    name = "package_metadata",
    attributes = [":license"],
    purl = (purl.builder()
        .type("github")
        .namespace("madler")
        .name("zlib")
        .version("v1.3.1")
        .build()),
    visibility = ["//visibility:public"],
)

license(
    name = "license",
    kind = "@package_metadata//licenses/spdx:Zlib",
    text = "LICENSE",
)

# ... the targets of the overlay.
```

Since you cannot add a `REPO.bazel` through `build_file`, use
`package(default_package_metadata = ...)` in each `BUILD` file of the overlay.

Keep the `version` in the PURL in sync with the archive URL. If the version is
already a variable in your `.bzl` file, generate the overlay with
`build_file_content` and compute the PURL there, as in
[Step 3](#step-3-annotate-repositories-created-by-repository-rules).


## Step 5 (optional): Write a rule that provides `PackageMetadataInfo`

Most rule sets can generate `package_metadata` targets as shown above. If your
rule set has its own rule that describes a package (e.g., an `npm_package` rule
that already knows the package name and version), you can make that rule
provide [`PackageMetadataInfo`](../metadata/providers/package_metadata_info.md#PackageMetadataInfo)
directly with
[`package_metadata_common.create_package_metadata`](../metadata/common/common.md#package_metadata_commoncreate_package_metadata).
This writes the metadata in the same format as `package_metadata`, so tools
treat both the same.

```starlark
# npm/npm_package_metadata.bzl
"""Declares rule `npm_package_metadata`."""

load("@package_metadata//common:common.bzl", "package_metadata_common")
load("@package_metadata//providers:package_attribute_info.bzl", "PackageAttributeInfo")
load("@package_metadata//providers:package_metadata_info.bzl", "PackageMetadataInfo")
load("@package_metadata//purl:purl.bzl", "purl")

def _npm_package_metadata_impl(ctx):
    builder = (purl.builder()
        .type("npm")
        .name(ctx.attr.package)
        .version(ctx.attr.version))
    if ctx.attr.scope:
        builder.namespace(ctx.attr.scope)

    info = package_metadata_common.create_package_metadata(
        actions = ctx.actions,
        label = ctx.label,
        purl = builder.build(),
        attributes = [a[PackageAttributeInfo] for a in ctx.attr.attributes],
    )

    return [
        info,
        DefaultInfo(
            files = depset(
                direct = [
                    info.metadata,
                ],
            ),
        ),
    ]

_npm_package_metadata = rule(
    implementation = _npm_package_metadata_impl,
    attrs = {
        "attributes": attr.label_list(
            providers = [
                PackageAttributeInfo,
            ],
        ),
        "package": attr.string(
            mandatory = True,
        ),
        "scope": attr.string(),
        "version": attr.string(
            mandatory = True,
        ),
    },
    provides = [
        PackageMetadataInfo,
    ],
)

def npm_package_metadata(
        *,
        name,
        package,
        version,
        scope = None,
        attributes = [],
        visibility = None,
        tags = None):
    _npm_package_metadata(
        name = name,
        package = package,
        version = version,
        scope = scope,
        attributes = attributes,
        visibility = visibility,
        tags = tags,
        # Required: without it, `default_package_metadata` applies to this
        # target itself, which is a dependency cycle.
        applicable_licenses = [],
    )
```

```starlark
npm_package_metadata(
    name = "package_metadata",
    scope = "@babel",
    package = "core",
    version = "7.26.0",
)
# purl: pkg:npm/%40babel/core@7.26.0
```

Two details matter here:

  - **Clear `applicable_licenses` on the metadata target.** A target that
    provides `PackageMetadataInfo` and is referenced from
    `default_package_metadata` must not itself pick up
    `default_package_metadata`, or analysis fails with a dependency cycle.
    Wrap the rule in a macro that passes `applicable_licenses = []` (the legacy
    name of `package_metadata`, which also works on Bazel 7), as all rules in
    `@package_metadata` do. The same applies to custom attribute rules that
    provide [`PackageAttributeInfo`](../metadata/providers/package_attribute_info.md#PackageAttributeInfo).

  - **Honor consumer overrides when reading metadata.** If your rule *reads*
    the metadata of other targets (e.g., to embed it in an artifact), resolve
    it with
    [`package_metadata_common.get_package_metadata`](../metadata/common/common.md#package_metadata_commonget_package_metadata)
    so that overrides registered on the
    [`package_metadata` toolchain](../metadata/rules/package_metadata_toolchain.md#package_metadata_toolchain)
    take effect.


## Step 6: Verify the metadata

Build the `package_metadata` target of a fetched repository and inspect the
generated JSON:

```sh
$ bazel build @guava//:package_metadata
$ cat "$(bazel cquery --output=files @guava//:package_metadata)"
{"attributes":{"build.bazel.attribute.license":"bazel-out/.../license.package-attribute.json"},"label":"@@+maven+guava//:package_metadata","purl":"pkg:maven/com.google.guava/guava@33.4.0-jre"}
```

Check that the targets of the repository are associated with it:

```sh
$ bazel query 'deps(@guava//:jar)'
@guava//:guava.jar
@guava//:jar
@guava//:license
@guava//:package_metadata
@package_metadata//licenses/spdx:Apache-2.0
```

To check a PURL in a test, [`purl.parse`](../metadata/purl/purl.md#purlparse) returns its
components, or an error if it is not a valid PURL.


## Checklist

  - [ ] `bazel_dep(name = "package_metadata", ...)` is a regular dependency of
        your module.
  - [ ] Your own module is annotated (see [As module author](../metadata/README.md#as-module-author)).
  - [ ] Every repository your rules create has a `package_metadata` target and
        sets `default_package_metadata` in `REPO.bazel` (or per package).
  - [ ] PURLs are built with `purl.builder` (or `purl.bazel`), using the
        ecosystem's PURL type and the exact resolved version.
  - [ ] `repository_url` is set for packages from non-default registries.
  - [ ] Licenses are declared when the package manager reports a single SPDX
        license ID.
  - [ ] Generated `BUILD` files load `@package_metadata` through
        `str(Label(...))`.
  - [ ] Custom rules providing `PackageMetadataInfo` or `PackageAttributeInfo`
        pass `applicable_licenses = []`.
