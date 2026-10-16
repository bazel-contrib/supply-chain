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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

const help = `  --[no]batch (a boolean; default: "false")
  --output_base (a path; default: see description)
  --compilation_mode [-c] (fastbuild, dbg or opt; default: "fastbuild")
  --[no]keep_going [-k] (a boolean; default: "false")
  --define (a 'name=value' assignment; may be used multiple times)
  --config (a string; may be used multiple times)
  --discard_analysis_cache
`

func TestCanonicalizeFullInvocation(t *testing.T) {
	var calls [][]string
	invoke := func(args []string) ([]byte, error) {
		calls = append(calls, append([]string{}, args...))
		if args[len(args)-1] == "--long" {
			return []byte(help), nil
		}
		return []byte("--compilation_mode=opt\n--keep_going\n--define=message=hello world 'quoted'\n"), nil
	}
	args := []string{"--output_base", "/tmp/output with spaces", "--batch", "run", "//app:main", "-c", "opt", "-k", "--define", "message=hello world 'quoted'", "--", "--program-flag", "two words", ""}
	got, err := canonicalize(args, invoke)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"--output_base", "/tmp/output with spaces", "--batch", "run", "--compilation_mode=opt", "--keep_going", "--define=message=hello world 'quoted'", "//app:main", "--", "--program-flag", "two words", ""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	wantCall := []string{"--output_base", "/tmp/output with spaces", "--batch", "canonicalize-flags", "--for_command=run", "--", "-c", "opt", "-k", "--define", "message=hello world 'quoted'"}
	if !reflect.DeepEqual(calls[len(calls)-1], wantCall) {
		t.Fatalf("canonicalizer received %q, want %q", calls[len(calls)-1], wantCall)
	}
}

func TestCanonicalizeTargetsAndEmptyOutput(t *testing.T) {
	for _, args := range [][]string{
		{"build", "//...", "--", "-//excluded/...", "--literal"},
		{"query", "deps(//app:main)"},
		{"build", "relative/target", "@repo//:target"},
	} {
		got, err := canonicalize(args, func(call []string) ([]byte, error) {
			if !reflect.DeepEqual(call, []string{"canonicalize-flags", "--for_command=" + args[0], "--"}) {
				t.Fatalf("unexpected call %q", call)
			}
			return nil, nil
		})
		if err != nil || !reflect.DeepEqual(got, args) {
			t.Fatalf("got %q, %v; want %q", got, err, args)
		}
	}
}

func TestCanonicalizeErrors(t *testing.T) {
	for _, args := range [][]string{
		{}, {"--batch"}, {"--output_base"}, {"build", "-c"}, {"build", "-x"},
		{"build", "--define=message=line one\nline two"},
	} {
		if _, err := canonicalize(args, func([]string) ([]byte, error) { return []byte(help), nil }); err == nil {
			t.Errorf("expected error for %q", args)
		}
	}
	failure := errors.New("invalid flag")
	if _, err := canonicalize([]string{"build", "--bad=value"}, func([]string) ([]byte, error) { return nil, failure }); !errors.Is(err, failure) {
		t.Fatalf("lost canonicalization error: %v", err)
	}
	if _, err := canonicalize([]string{"build"}, func([]string) ([]byte, error) { return []byte("unexpected output\n"), nil }); err == nil {
		t.Fatal("accepted malformed output")
	}
}

// The test executable doubles as a Bazel-compatible subprocess so these tests
// exercise actual argv boundaries, stdio, and exit status propagation.
func TestMain(m *testing.M) {
	if path := os.Getenv("RECORDING_SANDBOX_TEST_LOG"); path != "" {
		args := os.Args[1:]
		log, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		_ = json.NewEncoder(log).Encode(args)
		_ = log.Close()
		if args[0] == "canonicalize-flags" {
			if os.Getenv("RECORDING_SANDBOX_TEST_FAIL") != "" {
				fmt.Fprintln(os.Stderr, "invalid flag")
				os.Exit(2)
			}
			fmt.Println("--define=message=hello world")
			os.Exit(0)
		}
		data, _ := io.ReadAll(os.Stdin)
		fmt.Fprintf(os.Stdout, "stdin: %s", data)
		fmt.Fprintln(os.Stderr, "command stderr")
		os.Exit(7)
	}
	os.Exit(m.Run())
}

func TestRun(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			path := t.TempDir() + "/calls.jsonl"
			t.Setenv("RECORDING_SANDBOX_TEST_LOG", path)
			if fail {
				t.Setenv("RECORDING_SANDBOX_TEST_FAIL", "1")
			}
			var stdout, stderr bytes.Buffer
			code := run([]string{binary, "run", "--define=message=hello world", "//:app", "--", "hello world"}, strings.NewReader("input"), &stdout, &stderr)
			if fail {
				if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "invalid flag") {
					t.Fatalf("code %d, stdout %q, stderr %q", code, &stdout, &stderr)
				}
			} else if code != 7 || stdout.String() != "stdin: input" || stderr.String() != "command stderr\n" {
				t.Fatalf("code %d, stdout %q, stderr %q", code, &stdout, &stderr)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var calls [][]string
			decoder := json.NewDecoder(bytes.NewReader(data))
			for decoder.More() {
				var call []string
				if err := decoder.Decode(&call); err != nil {
					t.Fatal(err)
				}
				calls = append(calls, call)
			}
			if fail && len(calls) != 1 {
				t.Fatalf("executed original command after failure: %q", calls)
			}
			if !fail && (len(calls) != 2 || !reflect.DeepEqual(calls[1], []string{"run", "--define=message=hello world", "//:app", "--", "hello world"})) {
				t.Fatalf("unexpected invocations %q", calls)
			}
		})
	}
}

func TestUsage(t *testing.T) {
	var stderr bytes.Buffer
	if code := run(nil, nil, io.Discard, &stderr); code != 2 || !strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("got code %d, stderr %q", code, &stderr)
	}
}
