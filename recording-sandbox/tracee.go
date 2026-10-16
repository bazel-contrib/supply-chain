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
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bazel-contrib/supply-chain/recording-sandbox/internal/remote-apis-proxy"
)

type child struct {
	cmd             *exec.Cmd
	done            chan struct{}
	err             error // Read only after done is closed.
	failureReported bool
	startupHint     func() string
}

// Keep a bounded tail because Tracee's startup errors arrive on stderr, while
// its exit status alone cannot distinguish missing privileges from other errors.
type traceeStderr struct {
	io.Writer
	mu   sync.Mutex
	tail []byte
}

func (w *traceeStderr) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tail = append(w.tail, data...)
	if len(w.tail) > 8192 {
		w.tail = append([]byte(nil), w.tail[len(w.tail)-8192:]...)
	}
	return w.Writer.Write(data)
}

func (w *traceeStderr) hint() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if bytes.Contains(w.tail, []byte("could not set capabilities")) && bytes.Contains(w.tail, []byte("operation not permitted")) {
		return "; Tracee needs host-granted Linux capabilities for eBPF. Configure the host and execution workers to launch the tracer with the required capabilities: https://aquasecurity.github.io/tracee/v0.24/docs/install/prerequisites/"
	}
	return ""
}

// Both subprocesses can log concurrently, including when the caller supplies
// a writer such as bytes.Buffer that does not synchronize its own writes.
type lockedWriter struct {
	mu sync.Mutex
	io.Writer
}

func (w *lockedWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Writer.Write(data)
}

func startChild(cmd *exec.Cmd) (*child, error) {
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	p := &child{cmd: cmd, done: make(chan struct{})}
	go func() {
		p.err = cmd.Wait()
		close(p.done)
	}()
	return p, nil
}

// stop gives Tracee time to flush its event pipeline before resorting to SIGKILL.
func (p *child) stop() error {
	select {
	case <-p.done:
		return p.err
	default:
	}
	_ = p.cmd.Process.Signal(syscall.SIGTERM)
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	select {
	case <-p.done:
		return p.err
	case <-timer.C:
		_ = p.cmd.Process.Kill()
		<-p.done
		return fmt.Errorf("%s did not stop within 10 seconds", p.cmd.Path)
	}
}

// RECORDING_SANDBOX_TRACE_FILE overrides the temporary JSONL output file.
func runTraced(binary string, args []string, traceePath string, stdin io.Reader, stdout, stderr io.Writer, logActions bool) (code int) {
	stderr = &lockedWriter{Writer: stderr}
	var err error
	var output *os.File
	if path := os.Getenv("RECORDING_SANDBOX_TRACE_FILE"); path != "" {
		// Preserve an existing recording rather than silently overwriting it.
		output, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	} else {
		output, err = os.CreateTemp("", "recording-sandbox-*.jsonl")
	}
	if err != nil {
		fmt.Fprintln(stderr, "recording-sandbox: creating trace output:", err)
		return 1
	}
	defer output.Close()
	fmt.Fprintln(stderr, "recording-sandbox: trace output:", output.Name())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	trace, address, installDir, err := startTracee(traceePath, output, stderr)
	if err != nil {
		fmt.Fprintln(stderr, "recording-sandbox: starting Tracee:", err)
		return exitCode(err)
	}
	defer os.RemoveAll(installDir)
	defer func() {
		if err := trace.stop(); err != nil && !trace.failureReported {
			fmt.Fprintln(stderr, "recording-sandbox: stopping Tracee:", err)
			if code == 0 {
				code = 1
			}
		}
	}()
	if err := waitForTracee(trace, address, signals); err != nil {
		fmt.Fprintln(stderr, "recording-sandbox:", err)
		return 1
	}
	outputBase, err := os.MkdirTemp("", "recording-sandbox-bazel-*")
	if err != nil {
		fmt.Fprintln(stderr, "recording-sandbox:", err)
		return 1
	}
	session := &bazelSession{
		binary: binary, outputBase: outputBase, stderr: stderr,
		signals: signals, traceDone: trace.done,
	}
	var proxy *remoteapisproxy.Proxy
	// Shutdown must finish before Tracee is stopped so daemon termination and
	// the last build events are included in the recording, even on failure.
	defer func() {
		if proxy != nil {
			defer proxy.Close()
		}
		if err := session.close(); err != nil {
			fmt.Fprintln(stderr, "recording-sandbox:", err)
			if code == 0 {
				code = 1
			}
		}
	}()
	argv, err := canonicalize(args, session.invoke)
	if err != nil {
		fmt.Fprintln(stderr, "recording-sandbox:", err)
		return exitCode(err)
	}
	commandIndex := len(session.startup)
	switch argv[commandIndex] {
	case "build", "test", "run", "coverage":
		var actionLog io.Writer
		if logActions {
			actionLog = stderr
		}
		proxy, err = startProxy(traceePath, output.Name()+".actions", actionLog)
		if err != nil {
			fmt.Fprintln(stderr, "recording-sandbox: starting execution proxy:", err)
			return 1
		}
		var flags []string
		flags, _, err = remoteapisproxy.RewriteFlags(argv[commandIndex+1:], proxy.Endpoint())
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		argv = append(argv[:commandIndex+1], flags...)
	}
	cmd := exec.Command(binary, session.args(session.startup, argv[len(session.startup):]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	build, err := startChild(cmd)
	if err != nil {
		fmt.Fprintln(stderr, "recording-sandbox:", err)
		return exitCode(err)
	}
	for {
		select {
		case sig := <-signals:
			_ = build.cmd.Process.Signal(sig)
		case <-trace.done:
			// An incomplete recording must not be reported as a successful run.
			_ = build.stop()
			fmt.Fprintf(stderr, "recording-sandbox: Tracee exited during the build: %v\n", trace.err)
			trace.failureReported = true
			return 1
		case <-build.done:
			return exitCode(build.err)
		}
	}
}

func waitForTracee(trace *child, address string, signals <-chan os.Signal) error {
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: 250 * time.Millisecond}
	defer client.CloseIdleConnections()
	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-trace.done:
			trace.failureReported = true
			var hint string
			if trace.startupHint != nil {
				hint = trace.startupHint()
			}
			return fmt.Errorf("Tracee exited before becoming ready: %v%s", trace.err, hint)
		case sig := <-signals:
			return fmt.Errorf("interrupted while starting Tracee: %s", sig)
		case <-timeout.C:
			return fmt.Errorf("Tracee was not ready within 30 seconds")
		case <-tick.C:
			response, err := client.Get("http://" + address + "/healthz")
			if err == nil {
				_, _ = io.Copy(io.Discard, response.Body)
				_ = response.Body.Close()
				if response.StatusCode == http.StatusOK {
					return nil
				}
			}
		}
	}
}

// startTracee is shared by build recording and the action payload. Pdeathsig
// also stops the tracer if a remote worker forcibly kills the action wrapper.
func startTracee(traceePath string, output *os.File, stderr io.Writer) (*child, string, string, error) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, "", "", err
	}
	address := listener.Addr().String()
	_ = listener.Close()
	installDir, err := os.MkdirTemp("", "recording-sandbox-tracee-*")
	if err != nil {
		return nil, "", "", err
	}
	command := exec.Command(traceePath,
		"--scope", fmt.Sprintf("tree=%d", os.Getpid()),
		"--events", "syscalls,sched_process_exec,sched_process_fork,sched_process_exit",
		"--output", "json", "--server", "http-address="+address,
		"--server", "healthz", "--install-path", installDir,
	)
	diagnostics := &traceeStderr{Writer: stderr}
	command.Stdout, command.Stderr = output, diagnostics
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGTERM}
	trace, err := startChild(command)
	if err != nil {
		os.RemoveAll(installDir)
		return nil, "", "", err
	}
	trace.startupHint = diagnostics.hint
	return trace, address, installDir, nil
}
