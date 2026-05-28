#!/usr/bin/env bash
#
# Map external module name to where it is stored.

MODULE="$1"

declare -A module2dir
module2dir["package_metadata"]="metadata"
module2dir["purl"]="purl.bzl"
module2dir["supply-chain-go"]="lib/supplychain-go"
module2dir["supply_chain_tools"]="tools"
module2dir["supply_chain_examples"]="examples"

echo "${module2dir[${MODULE}]:-$MODULE}"
exit 0
