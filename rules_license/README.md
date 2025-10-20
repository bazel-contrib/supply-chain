# Compatiblity layer to make legacy rules_license declarations work with supply_chain


## Usage

This module does not appear in the BCR. You must load it with
an http_archive rule.

```
http_archive = use_repo_rule("@bazel_tools//build_defs/repo:http.bzl", "http_archive")

bazel_dep(name = "package_metadata", version = "0.0.5")
http_archive(
    name = "rules_license",
    # sha256 = TBD,
    strip_prefix = "supply-chain-dd_test/rules_license",
    urls = ["https://github.com/aiuto/supply-chain/archive/refs/tags/dd_test.tar.gz"],
)
```
