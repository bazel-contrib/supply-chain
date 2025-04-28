#!/usr/bin/env bash

set -o pipefail

bazel test //...
readonly exit_code="$?"

case "${exit_code}" in
  "4")
    # Status code indicates that the build succeeded but there were no tests to
    # run. Ignore and exit successfully.
    exit "0"
    ;;

  *)
    exit "${exit_code}"
    ;;
esac
