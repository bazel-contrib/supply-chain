#!/usr/bin/env python3
"""Regenerates the local registry in this directory.

registry/head serves every module of this repo (including rules_license, the
examples and external_packages) from its current directory (local_path), at
"HEAD", at the version declared in its MODULE.bazel, and at every version
published to the BCR (see PUBLISHED), always with the MODULE.bazel from this
repo. It replaces local_path_override and is used by the .bazelrc of each module.
Usage: python3 registry/generate.py [--refresh]
  --refresh re-reads the published version list from the BCR.
"""

import json
import re
import shutil
import sys
import urllib.request
from pathlib import Path

REGISTRY_DIR = Path(__file__).resolve().parent
REPO_ROOT = REGISTRY_DIR.parent

# module name -> directory in this repo.
MODULES = {
    "package_metadata": "metadata",
    "package_metadata_extensions": "metadata-extensions",
    "supply-chain-go": "lib/supplychain-go",
    "supply_chain_tools": "tools",
    "rules_license": "rules_license",
    "supply_chain_examples": "examples",
    **{
        p.name: f"external_packages/{p.name}"
        for p in sorted((REPO_ROOT / "external_packages").glob("*/MODULE.bazel"))
        for p in [p.parent]
    },
}

# Versions published to https://registry.bazel.build, per module.
PUBLISHED = {
    "package_metadata": [
        "0.0.1", "0.0.2", "0.0.3", "0.0.4", "0.0.5", "0.0.6", "0.0.7",
        "0.0.10", "0.0.13", "0.0.15",
    ],
    "supply-chain-go": ["0.0.4", "0.0.5", "0.0.6", "0.0.7", "0.0.10"],
    "supply_chain_tools": ["0.0.1"],
}

# Versions other modules request for a module that is replaced by a local copy
# (rules_license here is a modified fork without a version of its own).
EXTRA_VERSIONS = {"rules_license": ["1.0.0"]}

BCR_RAW = (
    "https://raw.githubusercontent.com/bazelbuild/bazel-central-registry/"
    "main/modules/{}"
)


def fetch(path):
    with urllib.request.urlopen(BCR_RAW.format(path)) as r:
        return r.read().decode()


def refresh_published():
    for name in PUBLISHED:
        PUBLISHED[name] = json.loads(fetch(f"{name}/metadata.json"))["versions"]


def module_file(directory, version):
    text = (REPO_ROOT / directory / "MODULE.bazel").read_text()
    # Registry entries must not carry the development-only overrides.
    text = re.sub(r"^local_path_override\(.*?^\)\n", "", text,
                  flags=re.S | re.M)
    text, n = re.subn(r'(module\([^)]*?version\s*=\s*)"[^"]*"',
                      rf'\1"{version}"', text, count=1, flags=re.S)
    if n == 0:
        # module() without a version (e.g. rules_license).
        text, n = re.subn(r'(module\(\s*name\s*=\s*"[^"]*",)',
                          rf'\1\n    version = "{version}",', text, count=1)
    assert n == 1, directory
    return text.rstrip() + "\n"


def declared_version(directory):
    text = (REPO_ROOT / directory / "MODULE.bazel").read_text()
    m = re.search(r'module\([^)]*?version\s*=\s*"([^"]*)"', text, re.S)
    return m.group(1) if m else ""


def head_versions():
    result = {}
    for module, directory in MODULES.items():
        versions = ["HEAD", declared_version(directory)]
        versions += PUBLISHED.get(module, []) + EXTRA_VERSIONS.get(module, [])
        result[module] = [v for i, v in enumerate(versions)
                          if v and v not in versions[:i]]
    return result


def write_registry(name, versions_by_module):
    root = REGISTRY_DIR / name
    shutil.rmtree(root, ignore_errors=True)
    (root).mkdir(parents=True)
    # registry/<name>/ -> repo root
    (root / "bazel_registry.json").write_text(
        json.dumps({"module_base_path": "../.."}, indent=2) + "\n")
    for module, versions in versions_by_module.items():
        directory = MODULES[module]
        mdir = root / "modules" / module
        mdir.mkdir(parents=True)
        (mdir / "metadata.json").write_text(
            json.dumps({"versions": versions, "yanked_versions": {}},
                       indent=2) + "\n")
        for v in versions:
            vdir = mdir / v
            vdir.mkdir()
            (vdir / "MODULE.bazel").write_text(module_file(directory, v))
            (vdir / "source.json").write_text(
                json.dumps({"type": "local_path", "path": directory},
                           indent=2) + "\n")


def main():
    if "--refresh" in sys.argv:
        refresh_published()
    write_registry("head", head_versions())


if __name__ == "__main__":
    main()
