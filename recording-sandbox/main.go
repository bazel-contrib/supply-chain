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
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: recording-sandbox BAZEL [STARTUP_OPTIONS...] COMMAND [ARGS...]")
		return 2
	}
	traceePath, err := bundledTracee()
	if err != nil {
		fmt.Fprintln(stderr, "recording-sandbox:", err)
		return 1
	}
	return runWithTracee(args, traceePath, stdin, stdout, stderr)
}

func runWithTracee(args []string, traceePath string, stdin io.Reader, stdout, stderr io.Writer) int {
	binary, err := exec.LookPath(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 127
	}
	return runTraced(binary, args[1:], traceePath, stdin, stdout, stderr)
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return 128 + int(status.Signal())
		}
		return exit.ExitCode()
	}
	return 1
}

type invokeFunc func(startup []string, commandArgs ...string) ([]byte, error)

func canonicalize(args []string, invoke invokeFunc) ([]string, error) {
	for _, arg := range args {
		if strings.ContainsAny(arg, "\r\n") {
			return nil, errors.New("arguments containing line breaks cannot be represented by canonicalize-flags")
		}
	}
	// Startup options precede the command. Read metadata lazily: options with
	// inline values need no help lookup, which also supports minimal Bazel-like
	// binaries implementing only the command and canonicalize-flags.
	startupEnd := 0
	var startupOptions map[string]bool
	for startupEnd < len(args) && strings.HasPrefix(args[startupEnd], "-") {
		arg := args[startupEnd]
		if !strings.Contains(arg, "=") {
			if startupOptions == nil {
				var err error
				startupOptions, err = loadOptions(invoke, args[:startupEnd], "startup_options")
				if err != nil {
					return nil, err
				}
			}
			value, known := startupOptions[arg]
			if !known {
				return nil, fmt.Errorf("unknown startup option %q", arg)
			}
			if value {
				startupEnd++
				if startupEnd == len(args) {
					return nil, fmt.Errorf("missing value for %s", arg)
				}
			}
		}
		startupEnd++
	}
	if startupEnd == len(args) {
		return nil, errors.New("missing Bazel command")
	}
	startup, command := args[:startupEnd], args[startupEnd]
	var flags, residue []string
	var commandOptions map[string]bool
	for i := startupEnd + 1; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			residue = append(residue, args[i:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			residue = append(residue, arg)
			continue
		}
		flags = append(flags, arg)
		if strings.Contains(arg, "=") {
			continue
		}
		if commandOptions == nil {
			var err error
			commandOptions, err = loadOptions(invoke, startup, command)
			if err != nil {
				return nil, err
			}
		}
		value, known := commandOptions[arg]
		if !known {
			// Starlark booleans and aliases are not listed in help. Bazel
			// accepts their values inline; leave validation to canonicalize-flags.
			if strings.HasPrefix(arg, "--") {
				continue
			}
			return nil, fmt.Errorf("unknown short option %q", arg)
		}
		if value {
			i++
			if i == len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			flags = append(flags, args[i])
		}
	}
	canonicalArgs := []string{"canonicalize-flags", "--for_command=" + command, "--"}
	canonicalArgs = append(canonicalArgs, flags...)
	output, err := invoke(startup, canonicalArgs...)
	if err != nil {
		return nil, fmt.Errorf("canonicalize-flags: %w", err)
	}
	result := append([]string{}, startup...)
	result = append(result, command)
	// Bazel emits one complete argument per line, not shell-quoted text.
	// Splitting on whitespace would corrupt values containing spaces or quotes.
	for _, flag := range strings.Split(strings.TrimSuffix(string(output), "\n"), "\n") {
		if flag == "" {
			continue
		}
		if !strings.HasPrefix(flag, "--") {
			return nil, fmt.Errorf("unexpected canonicalize-flags output %q", flag)
		}
		result = append(result, flag)
	}
	return append(result, residue...), nil
}

// Long help lists options as "--name [-x] (type; default: ...)". Boolean
// spellings use "--[no]name"; expansion flags have no parenthesized type.
var optionLine = regexp.MustCompile(`(?m)^  --(\[no\])?([a-zA-Z0-9_]+)(?: \[(-[a-zA-Z0-9])\])?(?: \(([^\n]*)\))?\r?$`)

func loadOptions(invoke invokeFunc, startup []string, command string) (map[string]bool, error) {
	output, err := invoke(startup, "help", command, "--long")
	if err != nil {
		return nil, fmt.Errorf("help %s: %w", command, err)
	}
	options := make(map[string]bool)
	for _, match := range optionLine.FindAllStringSubmatch(string(output), -1) {
		value := match[1] == "" && match[4] != ""
		options["--"+match[2]] = value
		if match[1] != "" {
			options["--no"+match[2]] = false
		}
		if match[3] != "" {
			options[match[3]] = value
		}
	}
	if len(options) == 0 {
		return nil, fmt.Errorf("help %s returned no option definitions", command)
	}
	return options, nil
}
