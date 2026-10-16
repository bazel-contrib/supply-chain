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
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
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
	invoke := func(startup []string, commandArgs ...string) ([]byte, error) {
		args := append(append([]string{}, startup...), commandArgs...)
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

// The test executable doubles as both Tracee and a Bazel-compatible binary.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "--recording-sandbox-test-daemon" {
		time.Sleep(time.Hour)
		os.Exit(0)
	}
	if path := os.Getenv("RECORDING_SANDBOX_TEST_LOG"); path != "" {
		args := os.Args[1:]
		log, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		_ = json.NewEncoder(log).Encode(args)
		_ = log.Close()
		if args[0] == "--scope" {
			fakeTracee(args)
			return
		}
		command := ""
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				command = arg
				break
			}
		}
		switch command {
		case "canonicalize-flags":
			if os.Getenv("RECORDING_SANDBOX_TEST_FAIL") != "" {
				fmt.Fprintln(os.Stderr, "invalid flag")
				os.Exit(2)
			}
			fmt.Println("--define=message=hello world")
			os.Exit(0)
		case "shutdown":
			if os.Getenv("RECORDING_SANDBOX_TEST_SHUTDOWN_FAIL") != "" {
				os.Exit(9)
			}
			_ = os.WriteFile(os.Getenv("RECORDING_SANDBOX_TRACE_FILE")+".shutdown", nil, 0600)
			os.Exit(0)
		default:
			data, _ := io.ReadAll(os.Stdin)
			fmt.Fprintf(os.Stdout, "stdin: %s", data)
			fmt.Fprintln(os.Stderr, "command stderr")
			if os.Getenv("RECORDING_SANDBOX_TEST_COMMAND_FAIL") != "" {
				os.Exit(7)
			}
			os.Exit(0)
		}
	}
	os.Exit(m.Run())
}

func TestDaemonKilledAfterShutdown(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("RECORDING_SANDBOX_TEST_LOG", filepath.Join(root, "calls"))
			t.Setenv("RECORDING_SANDBOX_TRACE_FILE", filepath.Join(root, "trace"))
			if fail {
				t.Setenv("RECORDING_SANDBOX_TEST_SHUTDOWN_FAIL", "1")
			}
			daemon := exec.Command(binary, "--recording-sandbox-test-daemon")
			if err := daemon.Start(); err != nil {
				t.Fatal(err)
			}
			defer daemon.Wait()
			defer daemon.Process.Kill()
			outputBase := filepath.Join(root, "output")
			if err := os.MkdirAll(filepath.Join(outputBase, "server"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(outputBase, "server", "server.pid.txt"), []byte(fmt.Sprint(daemon.Process.Pid)), 0600); err != nil {
				t.Fatal(err)
			}
			if processStartTime(daemon.Process.Pid) == "" {
				t.Fatal("daemon was not running")
			}
			session := &bazelSession{binary: binary, outputBase: outputBase, stderr: io.Discard}
			err := session.close()
			if (err != nil) != fail {
				t.Fatalf("shutdown error %v, want failure %v", err, fail)
			}
			if processStartTime(daemon.Process.Pid) != "" {
				t.Fatal("Bazel daemon survived invocation cleanup")
			}
		})
	}
}

func fakeTracee(args []string) {
	if os.Getenv("RECORDING_SANDBOX_TEST_TRACEE_FAIL") != "" {
		fmt.Fprintln(os.Stderr, os.Getenv("RECORDING_SANDBOX_TEST_TRACEE_ERROR"))
		os.Exit(9)
	}
	var address string
	for _, arg := range args {
		if strings.HasPrefix(arg, "http-address=") {
			address = strings.TrimPrefix(arg, "http-address=")
		}
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	fmt.Println(`{"ready":true}`)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})
	go http.Serve(listener, mux)
	<-signals
	if _, err := os.Stat(os.Getenv("RECORDING_SANDBOX_TRACE_FILE") + ".shutdown"); err != nil && os.Getenv("RECORDING_SANDBOX_TEST_ACTION") == "" {
		os.Exit(10) // Bazel must shut down before Tracee is stopped.
	}
	fmt.Println(`{"flushed":true}`)
	os.Exit(0)
}

func TestActionMode(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	t.Setenv("RECORDING_SANDBOX_TEST_LOG", filepath.Join(root, "calls.jsonl"))
	t.Setenv("RECORDING_SANDBOX_TEST_ACTION", "1")
	tracePath := filepath.Join(root, "outputs", "trace.jsonl")
	var stdout, stderr bytes.Buffer
	code := run([]string{"--action", "--tracee=" + binary, "--trace-output=" + tracePath, "--", binary, "payload", "two words"}, strings.NewReader("action input"), &stdout, &stderr)
	if code != 0 || stdout.String() != "stdin: action input" || stderr.String() != "command stderr\n" {
		t.Fatalf("action exit %d stdout %q stderr %q", code, &stdout, &stderr)
	}
	trace, err := os.ReadFile(tracePath)
	if err != nil || string(trace) != "{\"ready\":true}\n{\"flushed\":true}\n" {
		t.Fatalf("action trace %s: %v", trace, err)
	}
	calls := readCalls(t, filepath.Join(root, "calls.jsonl"))
	if len(calls) != 2 || calls[0][0] != "--scope" || !reflect.DeepEqual(calls[1], []string{"payload", "two words"}) {
		t.Fatalf("action was treated as a Bazel invocation: %q", calls)
	}
}

func TestRun(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name                        string
		flagFailure, commandFailure bool
		code                        int
	}{
		{name: "success", code: 0},
		{name: "command failure", commandFailure: true, code: 7},
		{name: "canonicalization failure", flagFailure: true, code: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir() + "/calls.jsonl"
			t.Setenv("RECORDING_SANDBOX_TEST_LOG", path)
			tracePath := t.TempDir() + "/trace.jsonl"
			t.Setenv("RECORDING_SANDBOX_TRACE_FILE", tracePath)
			if test.flagFailure {
				t.Setenv("RECORDING_SANDBOX_TEST_FAIL", "1")
			}
			if test.commandFailure {
				t.Setenv("RECORDING_SANDBOX_TEST_COMMAND_FAIL", "1")
			}
			var stdout, stderr bytes.Buffer
			code := runWithTracee([]string{binary, "--output_base=/user/output", "run", "--define=message=hello world", "//:app", "--", "hello world"}, binary, strings.NewReader("input"), &stdout, &stderr, false)
			if code != test.code {
				t.Fatalf("code %d, want %d; stderr %q", code, test.code, &stderr)
			}
			if test.flagFailure {
				if stdout.Len() != 0 || !strings.Contains(stderr.String(), "invalid flag") {
					t.Fatalf("stdout %q, stderr %q", &stdout, &stderr)
				}
			} else if stdout.String() != "stdin: input" || !strings.Contains(stderr.String(), "command stderr\n") {
				t.Fatalf("stdout %q, stderr %q", &stdout, &stderr)
			}
			calls := readCalls(t, path)
			wantCommands := []string{"canonicalize-flags", "run", "shutdown"}
			if test.flagFailure {
				wantCommands = []string{"canonicalize-flags", "shutdown"}
			}
			if len(calls) != len(wantCommands)+1 || calls[0][0] != "--scope" {
				t.Fatalf("unexpected calls: %q", calls)
			}
			var outputBase string
			for i, call := range calls[1:] {
				base := ""
				for _, arg := range call {
					if arg == "--batch" || strings.HasPrefix(arg, "--batch=") {
						t.Fatalf("legacy batch mode: %q", call)
					}
					if strings.HasPrefix(arg, "--output_base=") {
						base = strings.TrimPrefix(arg, "--output_base=")
					}
				}
				if base == "" || base == "/user/output" || (i > 0 && base != outputBase) {
					t.Fatalf("commands did not share a private daemon: %q", calls)
				}
				outputBase = base
				if !strings.Contains(strings.Join(call, " "), "--nobatch --max_idle_secs=30 "+wantCommands[i]) {
					t.Fatalf("unexpected command: %q", call)
				}
			}
			if _, err := os.Stat(outputBase); !os.IsNotExist(err) {
				t.Fatalf("private output base not removed: %v", err)
			}
			trace, err := os.ReadFile(tracePath)
			if err != nil || string(trace) != "{\"ready\":true}\n{\"flushed\":true}\n" {
				t.Fatalf("trace was not flushed after daemon shutdown: %s, %v", trace, err)
			}
		})
	}
}

func readCalls(t *testing.T, path string) [][]string {
	t.Helper()
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
	return calls
}

func TestTraceeStartupFailure(t *testing.T) {
	binary, _ := os.Executable()
	path := t.TempDir() + "/calls.jsonl"
	t.Setenv("RECORDING_SANDBOX_TEST_LOG", path)
	t.Setenv("RECORDING_SANDBOX_TRACE_FILE", t.TempDir()+"/trace.jsonl")
	t.Setenv("RECORDING_SANDBOX_TEST_TRACEE_FAIL", "1")
	var stderr bytes.Buffer
	if code := runWithTracee([]string{binary, "build"}, binary, nil, io.Discard, &stderr, false); code != 1 || !strings.Contains(stderr.String(), "Tracee exited before becoming ready") {
		t.Fatalf("code %d, stderr %q", code, &stderr)
	}
	if strings.Contains(stderr.String(), "stopping Tracee:") || strings.Contains(stderr.String(), "host-granted") {
		t.Fatalf("duplicate or misleading diagnostic: %q", &stderr)
	}
	if calls := readCalls(t, path); len(calls) != 1 || calls[0][0] != "--scope" {
		t.Fatalf("Bazel executed despite Tracee failing to start: %q", calls)
	}
}

func TestTraceeCapabilityFailure(t *testing.T) {
	binary, _ := os.Executable()
	path := t.TempDir() + "/calls.jsonl"
	t.Setenv("RECORDING_SANDBOX_TEST_LOG", path)
	t.Setenv("RECORDING_SANDBOX_TEST_TRACEE_FAIL", "1")
	t.Setenv("RECORDING_SANDBOX_TEST_TRACEE_ERROR", `{"level":"fatal","error":"capabilities.(*Capabilities).EBPF: could not set capabilities: operation not permitted"}`)
	for _, action := range []bool{false, true} {
		t.Run(fmt.Sprint("action=", action), func(t *testing.T) {
			tracePath := t.TempDir() + "/trace.jsonl"
			t.Setenv("RECORDING_SANDBOX_TRACE_FILE", tracePath)
			var stderr bytes.Buffer
			var code int
			if action {
				code = run([]string{"--action", "--tracee=" + binary, "--trace-output=" + tracePath, "--", binary, "payload"}, nil, io.Discard, &stderr)
			} else {
				code = runWithTracee([]string{binary, "build"}, binary, nil, io.Discard, &stderr, false)
			}
			if code != 1 || !strings.Contains(stderr.String(), "host-granted Linux capabilities") || strings.Contains(stderr.String(), "stopping Tracee:") {
				t.Fatalf("code %d, stderr %q", code, &stderr)
			}
		})
	}
	for _, call := range readCalls(t, path) {
		if call[0] != "--scope" {
			t.Fatalf("command executed despite Tracee capability failure: %q", call)
		}
	}
}

func TestBundledTracee(t *testing.T) {
	if os.Getenv("TEST_SRCDIR") == "" {
		t.Skip("release bundle is supplied by Bazel runfiles")
	}
	path, err := bundledTracee()
	if err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(path, "version").CombinedOutput()
	if err != nil || !strings.Contains(string(output), "0.24.1") {
		t.Fatalf("bundled Tracee: %s, %v", output, err)
	}
}

func TestUsage(t *testing.T) {
	var stderr bytes.Buffer
	if code := run(nil, nil, io.Discard, &stderr); code != 2 || !strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("got code %d, stderr %q", code, &stderr)
	}
}

func TestCanonicalizeTargetsAndEmptyOutput(t *testing.T) {
	for _, args := range [][]string{
		{"build", "//...", "--", "-//excluded/...", "--literal"},
		{"query", "deps(//app:main)"},
		{"build", "relative/target", "@repo//:target"},
	} {
		got, err := canonicalize(args, func(startup []string, call ...string) ([]byte, error) {
			if !reflect.DeepEqual(call, []string{"canonicalize-flags", "--for_command=" + args[0], "--"}) {
				t.Fatalf("unexpected call %q", call)
			}
			return nil, nil
		})
		want := args
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("got %q, %v; want %q", got, err, want)
		}
	}
}

func TestCanonicalizeErrors(t *testing.T) {
	for _, args := range [][]string{
		{}, {"--batch"}, {"--output_base"}, {"build", "-c"}, {"build", "-x"},
		{"build", "--define=message=line one\nline two"},
	} {
		if _, err := canonicalize(args, func([]string, ...string) ([]byte, error) { return []byte(help), nil }); err == nil {
			t.Errorf("expected error for %q", args)
		}
	}
	failure := errors.New("invalid flag")
	if _, err := canonicalize([]string{"build", "--bad=value"}, func([]string, ...string) ([]byte, error) { return nil, failure }); !errors.Is(err, failure) {
		t.Fatalf("lost canonicalization error: %v", err)
	}
	if _, err := canonicalize([]string{"build"}, func([]string, ...string) ([]byte, error) { return []byte("unexpected output\n"), nil }); err == nil {
		t.Fatal("accepted malformed output")
	}
}
