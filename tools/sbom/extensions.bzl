"""Module extensions for SBOM validation tools."""

_CYCLONEDX_CLI_VERSION = "v0.33.1"
_SPDX_TOOLS_JAVA_VERSION = "v2.0.7"

_CYCLONEDX_CLI_PLATFORMS = {
    "linux_arm": struct(asset = "cyclonedx-linux-arm", constraints = ["@platforms//os:linux", "@platforms//cpu:armv7"]),
    "linux_arm64": struct(asset = "cyclonedx-linux-arm64", constraints = ["@platforms//os:linux", "@platforms//cpu:aarch64"]),
    "linux_x64": struct(asset = "cyclonedx-linux-x64", constraints = ["@platforms//os:linux", "@platforms//cpu:x86_64"]),
    "osx_arm64": struct(asset = "cyclonedx-osx-arm64", constraints = ["@platforms//os:macos", "@platforms//cpu:aarch64"]),
    "osx_x64": struct(asset = "cyclonedx-osx-x64", constraints = ["@platforms//os:macos", "@platforms//cpu:x86_64"]),
    "win_arm64": struct(asset = "cyclonedx-win-arm64.exe", constraints = ["@platforms//os:windows", "@platforms//cpu:aarch64"]),
    "win_x64": struct(asset = "cyclonedx-win-x64.exe", constraints = ["@platforms//os:windows", "@platforms//cpu:x86_64"]),
    "win_x86": struct(asset = "cyclonedx-win-x86.exe", constraints = ["@platforms//os:windows", "@platforms//cpu:x86_32"]),
}

def _maybe_sha256(kwargs, sha256):
    if sha256:
        kwargs["sha256"] = sha256
    return kwargs

def _cyclonedx_cli_repository_impl(repository_ctx):
    output = "cyclonedx.exe" if repository_ctx.attr.asset.endswith(".exe") else "cyclonedx"
    url = repository_ctx.attr.url_template.format(
        asset = repository_ctx.attr.asset,
        version = repository_ctx.attr.version,
    )
    repository_ctx.download(
        output = output,
        executable = True,
        **_maybe_sha256({"url": url}, repository_ctx.attr.sha256)
    )
    repository_ctx.file("BUILD.bazel", """package(default_visibility = ["//visibility:public"])

exports_files(["%s"])

load("@supply_chain_tools//sbom:validators.bzl", "cyclonedx_validator_toolchain")

cyclonedx_validator_toolchain(
    name = "toolchain_impl",
    validator = ":%s",
)

toolchain(
    name = "toolchain",
    exec_compatible_with = %s,
    toolchain = ":toolchain_impl",
    toolchain_type = "@supply_chain_tools//sbom:cyclonedx_validator_toolchain_type",
)
""" % (output, output, str(repository_ctx.attr.constraints)))

cyclonedx_cli_repository = repository_rule(
    implementation = _cyclonedx_cli_repository_impl,
    attrs = {
        "asset": attr.string(mandatory = True, doc = "CycloneDX CLI release asset name."),
        "constraints": attr.string_list(mandatory = True, doc = "Exec platform constraints for this asset."),
        "sha256": attr.string(doc = "Optional SHA-256 checksum for the selected CycloneDX CLI asset."),
        "url_template": attr.string(default = "https://github.com/CycloneDX/cyclonedx-cli/releases/download/{version}/{asset}", doc = "URL template for CycloneDX CLI assets. Supports {version} and {asset}."),
        "version": attr.string(default = _CYCLONEDX_CLI_VERSION, doc = "CycloneDX CLI release tag."),
    },
)

def _spdx_tools_java_repository_impl(repository_ctx):
    version = repository_ctx.attr.version
    version_without_v = version[1:] if version.startswith("v") else version
    archive = "tools-java-%s.zip" % version_without_v
    jar = "tools-java-%s-jar-with-dependencies.jar" % version_without_v
    url = repository_ctx.attr.url or "https://github.com/spdx/tools-java/releases/download/%s/%s" % (version, archive)
    repository_ctx.download_and_extract(
        **_maybe_sha256({"url": url}, repository_ctx.attr.sha256)
    )
    repository_ctx.file("spdx-validate.sh", """#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -ne 1 ]]; then
    echo "Usage: $0 <spdx-file>" >&2
    exit 2
fi

# --- begin runfiles.bash initialization v3 ---
# Copy-pasted from the Bazel Bash runfiles library v3.
set -uo pipefail; set +e; f=bazel_tools/tools/bash/runfiles/runfiles.bash
# shellcheck disable=SC1090
source "${RUNFILES_DIR:-/dev/null}/$f" 2>/dev/null || \\
  source "$(grep -sm1 "^$f " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null || \\
  source "$0.runfiles/$f" 2>/dev/null || \\
  source "$(grep -sm1 "^$f " "$0.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \\
  source "$(grep -sm1 "^$f " "$0.exe.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \\
  { echo>&2 "ERROR: cannot find $f"; exit 1; }; f=; set -e
# --- end runfiles.bash initialization v3 ---

jar="$(rlocation "spdx_tools_java/%s")"
if [[ -z "$jar" ]]; then
    echo "ERROR: cannot find SPDX tools-java jar in runfiles" >&2
    exit 1
fi

exec java -Dorg.spdx.useJARLicenseInfoOnly=true -jar "$jar" Verify "$1"
""" % jar, executable = True)
    repository_ctx.file("BUILD.bazel", """package(default_visibility = ["//visibility:public"])

load("@rules_shell//shell:sh_binary.bzl", "sh_binary")

filegroup(
    name = "tools-java-jar",
    srcs = ["%s"],
)

sh_binary(
    name = "validate",
    srcs = ["spdx-validate.sh"],
    data = [":tools-java-jar"],
    deps = ["@rules_shell//shell/runfiles"],
)

load("@supply_chain_tools//sbom:validators.bzl", "spdx_validator_toolchain")

spdx_validator_toolchain(
    name = "toolchain_impl",
    validator = ":validate",
)

toolchain(
    name = "toolchain",
    toolchain = ":toolchain_impl",
    toolchain_type = "@supply_chain_tools//sbom:spdx_validator_toolchain_type",
)
""" % jar)

spdx_tools_java_repository = repository_rule(
    implementation = _spdx_tools_java_repository_impl,
    attrs = {
        "sha256": attr.string(doc = "Optional SHA-256 checksum for the SPDX tools-java release zip."),
        "url": attr.string(doc = "Optional full URL for the SPDX tools-java release zip."),
        "version": attr.string(default = _SPDX_TOOLS_JAVA_VERSION, doc = "SPDX tools-java release tag."),
    },
)

_cyclonedx_cli_tag = tag_class(attrs = {
    "sha256s": attr.string_dict(doc = "Optional SHA-256 checksums keyed by CycloneDX CLI asset name."),
    "url_template": attr.string(default = "https://github.com/CycloneDX/cyclonedx-cli/releases/download/{version}/{asset}", doc = "URL template for CycloneDX CLI assets. Supports {version} and {asset}."),
    "version": attr.string(default = _CYCLONEDX_CLI_VERSION, doc = "CycloneDX CLI release tag."),
})

_spdx_tools_java_tag = tag_class(attrs = {
    "sha256": attr.string(doc = "Optional SHA-256 checksum for the SPDX tools-java release zip."),
    "url": attr.string(doc = "Optional full URL for the SPDX tools-java release zip."),
    "version": attr.string(default = _SPDX_TOOLS_JAVA_VERSION, doc = "SPDX tools-java release tag."),
})

def _select_cyclonedx_tag(module_ctx):
    selected = {
        "sha256s": {},
        "url_template": "https://github.com/CycloneDX/cyclonedx-cli/releases/download/{version}/{asset}",
        "version": _CYCLONEDX_CLI_VERSION,
    }
    for module in module_ctx.modules:
        if not module.tags.cyclonedx_cli:
            continue
        tag = module.tags.cyclonedx_cli[-1]
        values = {
            "sha256s": tag.sha256s,
            "url_template": tag.url_template,
            "version": tag.version,
        }
        if module.is_root:
            return values
        selected = values
    return selected

def _select_spdx_tag(module_ctx, defaults):
    selected = defaults
    for module in module_ctx.modules:
        if not module.tags.spdx_tools_java:
            continue
        tag = module.tags.spdx_tools_java[-1]
        values = {
            "sha256": tag.sha256,
            "url": tag.url,
            "version": tag.version,
        }
        if module.is_root:
            return values
        selected = values
    return selected

def _sbom_validators_impl(module_ctx):
    cyclonedx = _select_cyclonedx_tag(module_ctx)
    spdx = _select_spdx_tag(module_ctx, {
        "sha256": "",
        "url": "",
        "version": _SPDX_TOOLS_JAVA_VERSION,
    })
    for platform, info in _CYCLONEDX_CLI_PLATFORMS.items():
        cyclonedx_cli_repository(
            name = "cyclonedx_cli_%s" % platform,
            asset = info.asset,
            constraints = info.constraints,
            sha256 = cyclonedx["sha256s"].get(info.asset, ""),
            url_template = cyclonedx["url_template"],
            version = cyclonedx["version"],
        )
    spdx_tools_java_repository(name = "spdx_tools_java", **spdx)

sbom_validators = module_extension(
    implementation = _sbom_validators_impl,
    tag_classes = {
        "cyclonedx_cli": _cyclonedx_cli_tag,
        "spdx_tools_java": _spdx_tools_java_tag,
    },
)
