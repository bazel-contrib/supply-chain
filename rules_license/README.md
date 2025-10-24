# Compatiblity layer to make legacy rules_license declarations work with supply_chain


## Usage

This module does not appear in the BCR. You must load it with
an http_archive rule.

```
http_archive = use_repo_rule("@bazel_tools//build_defs/repo:http.bzl", "http_archive")

bazel_dep(name = "package_metadata", version = "0.0.6")
bazel_dep(name = "rules_license", version = "1.0.0")
archive_override(
    module_name = "rules_license",
    sha256 = ...,
    strip_prefix = "supply-chain-0.0.6/rules_license",
    urls = ["https://github.com/bazel-contrib/supply-chain/archive/refs/tags/v0.0.6.tar.gz"],
)
```
