#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

# Passed as argument when invoking the script. E.g supply_chain_tools-1.2.3
TAG="${1}"

case "${TAG}" in 
  v[0-9]* )
    # The prefix is chosen to match what GitHub generates for source archives
    # This guarantees that users can easily switch from a released artifact to a source archive
    # with minimal differences in their code (e.g. strip_prefix remains the same)
    VERSION="${TAG:1}"
    PREFIX="supply-chain-${TAG:1}"
    ARCHIVE="supply-chain-$TAG.tar.gz"
    MODULE="package_metadata"
    STRIP_PREFIX="${PREFIX}/metadata",
    ;;
  * )
    VERSION=$(echo "$TAG" | sed -e 's/^.*-//')
    MODULE=$(echo "$TAG" | sed -e 's/-[0-9]*.*$//')
    PREFIX="${TAG}"
    ARCHIVE="${TAG}.tar.gz"
    STRIP_PREFIX="${PREFIX}",
esac

echo TAG=$TAG
echo VERSION=$VERSION
echo ARCHIVE=$ARCHIVE
echo MODULE=$MODULE

# NB: configuration for 'git archive' is in /.gitattributes
echo git archive --format=tar --prefix=${PREFIX}/ ${TAG} | gzip > $ARCHIVE
git archive --format=tar --prefix=${PREFIX}/ head | gzip > $ARCHIVE
SHA=$(shasum -a 256 $ARCHIVE | awk '{print $1}')

cat << EOF
## Using Bzlmod with Bazel 6 or greater

1. (Bazel 6 only) Enable with \`common --enable_bzlmod\` in \`.bazelrc\`.
2. Add to your \`MODULE.bazel\` file:

\`\`\`starlark
bazel_dep(name = "${MODULE}", version = "${VERSION}")
\`\`\`

## Using WORKSPACE

Paste this snippet into your \`WORKSPACE.bazel\` file:

\`\`\`starlark
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
    name = "${MODULE}",
    sha256 = "${SHA}",
    strip_prefix = "${STRIP_PREFIX}",
    url = "https://github.com/bazel-contrib/supply-chain/releases/download/${TAG}/${ARCHIVE}",
)
\`\`\`
EOF
