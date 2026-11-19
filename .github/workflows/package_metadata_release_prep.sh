#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

# Passed as argument when invoking the script.
TAG="${1}"

# The prefix is chosen to match what GitHub generates for source archives
# This guarantees that users can easily switch from a released artifact to a source archive
# with minimal differences in their code (e.g. strip_prefix remains the same)
PREFIX="package-metadata-${TAG:1}"
ARCHIVE="package-metadata-$TAG.tar.gz"

# Create a temporary directory for the package_metadata content
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Copy the metadata directory to the temporary location with the correct prefix
mkdir -p "$TEMP_DIR/$PREFIX"
cp -r metadata/* "$TEMP_DIR/$PREFIX/"

# Create the archive from the temporary directory
cd "$TEMP_DIR"
tar -czf "$OLDPWD/$ARCHIVE" "$PREFIX"
cd "$OLDPWD"

SHA=$(shasum -a 256 $ARCHIVE | awk '{print $1}')

cat << EOF
## Using Bzlmod with Bazel 6 or greater

1. (Bazel 6 only) Enable with \`common --enable_bzlmod\` in \`.bazelrc\`.
2. Add to your \`MODULE.bazel\` file:

\`\`\`starlark
bazel_dep(name = "package_metadata", version = "${TAG:1}")
\`\`\`

## Using WORKSPACE

Paste this snippet into your \`WORKSPACE.bazel\` file:

\`\`\`starlark
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
    name = "package_metadata",
    sha256 = "${SHA}",
    strip_prefix = "${PREFIX}",
    url = "https://github.com/bazel-contrib/supply-chain/releases/download/${TAG}/${ARCHIVE}",
)
\`\`\`
EOF
