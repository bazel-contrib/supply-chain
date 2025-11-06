# Compatiblity layer to make legacy rules_license declarations work with supply_chain


## Usage

This module does not appear in the BCR. You must load it with
an archive_override rule.

```
bazel_dep(name = "package_metadata", version = "0.0.6")
bazel_dep(name = "rules_license", version = "1.0.0")
archive_override(
    module_name = "rules_license",
    # 
    sha256 = "5bd0cc7594ea528fd28f98d82457f157827d48cc20e07bcfdbb56072f35c8f67",
    strip_prefix = "supply-chain-0.0.6/rules_license",
    urls = ["https://github.com/bazel-contrib/supply-chain/releases/download/v0.0.6/supply-chain-v0.0.6.tar.gz"],
)
```
