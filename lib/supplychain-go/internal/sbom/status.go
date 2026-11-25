package sbom

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ReadBuildTimestamp reads a BUILD_TIMESTAMP key out of a Bazel
// volatile-status.txt file (see ctx.version_file / --workspace_status_command),
// for use as the creation time of a generated SBOM. If path is empty, or the
// file has no BUILD_TIMESTAMP key, it returns the Unix epoch: the same fixed
// value Bazel itself substitutes for BUILD_TIMESTAMP when building without
// --stamp, so SBOM generation stays deterministic and cacheable unless the
// caller explicitly opts into a real timestamp via --stamp.
func ReadBuildTimestamp(path string) (time.Time, error) {
	if path == "" {
		return time.Unix(0, 0).UTC(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		key, value, found := strings.Cut(line, " ")
		if !found || key != "BUILD_TIMESTAMP" {
			continue
		}
		seconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("parsing BUILD_TIMESTAMP %q: %w", value, err)
		}
		return time.Unix(seconds, 0).UTC(), nil
	}

	return time.Unix(0, 0).UTC(), nil
}
