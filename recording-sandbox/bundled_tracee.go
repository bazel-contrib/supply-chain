// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"runtime"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

func bundledTracee() (string, error) {
	if runtime.GOOS != "linux" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		return "", fmt.Errorf("Tracee supports Linux amd64 and arm64; got %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	path, err := runfiles.Rlocation("tracee_linux_" + runtime.GOARCH + "/dist/tracee-static")
	if err != nil {
		return "", fmt.Errorf("locating bundled Tracee (build with Bazel): %w", err)
	}
	return path, nil
}
