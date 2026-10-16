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
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Each recording owns an output base and therefore a fresh Bazel server. All
// commands, including help and canonicalization, use that same server.
type bazelSession struct {
	binary     string
	outputBase string
	startup    []string
	stderr     io.Writer
	signals    <-chan os.Signal
	traceDone  <-chan struct{}
}

func (s *bazelSession) args(startup []string, commandArgs ...string) []string {
	args := append([]string{}, startup...)
	args = append(args, "--output_base="+s.outputBase, "--nobatch", "--max_idle_secs=30")
	return append(args, commandArgs...)
}

func (s *bazelSession) invoke(startup []string, commandArgs ...string) ([]byte, error) {
	s.startup = append([]string{}, startup...)
	cmd := exec.Command(s.binary, s.args(startup, commandArgs...)...)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, s.stderr
	process, err := startChild(cmd)
	if err != nil {
		return nil, err
	}
	for {
		select {
		case <-process.done:
			return output.Bytes(), process.err
		case sig := <-s.signals:
			_ = process.cmd.Process.Signal(sig)
		case <-s.traceDone:
			_ = process.stop()
			return nil, fmt.Errorf("Tracee exited during Bazel preparation")
		}
	}
}

func (s *bazelSession) close() error {
	// Remember the server identity before asking it to stop. If shutdown fails,
	// only kill this invocation's daemon, after checking its /proc start time.
	pidData, _ := os.ReadFile(filepath.Join(s.outputBase, "server", "server.pid.txt"))
	pid, _ := strconv.Atoi(strings.TrimSpace(string(pidData)))
	identity := processStartTime(pid)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.binary, s.args(s.startup, "shutdown")...)
	var diagnostic bytes.Buffer
	cmd.Stdout, cmd.Stderr = &diagnostic, &diagnostic
	err := cmd.Run()
	if pid > 1 && identity != "" && processStartTime(pid) == identity {
		if killErr := syscall.Kill(pid, syscall.SIGKILL); killErr != nil && killErr != syscall.ESRCH {
			return fmt.Errorf("killing Bazel daemon %d: %w", pid, killErr)
		}
		deadline := time.Now().Add(5 * time.Second)
		for processStartTime(pid) == identity {
			if time.Now().After(deadline) {
				return fmt.Errorf("Bazel daemon %d did not exit", pid)
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	if err != nil {
		return fmt.Errorf("Bazel shutdown: %w: %s", err, diagnostic.String())
	}
	return os.RemoveAll(s.outputBase)
}

// Linux stat's starttime identifies a process despite PID reuse. Zombies have
// already exited and cannot perform further build work.
func processStartTime(pid int) string {
	if pid <= 1 {
		return ""
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return ""
	}
	end := bytes.LastIndexByte(data, ')')
	if end < 0 {
		return ""
	}
	fields := strings.Fields(string(data[end+1:]))
	if len(fields) < 20 || fields[0] == "Z" {
		return ""
	}
	return fields[19]
}
