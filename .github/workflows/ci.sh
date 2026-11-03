#!/usr/bin/env bash

set -o pipefail

# Target patterns to test, space-separated. Defaults to everything in the
# current module. The `tests` sub-modules additionally pass the module under
# test (e.g. `@package_metadata//...`) so it gets built from an external
# standpoint, like consumers and the BCR presubmit do.
read -r -a targets <<< "${TARGETS:-//...}"

bazel test "${targets[@]}"
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
