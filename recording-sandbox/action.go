// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

// This entry point is a CAS payload for a single action, not another Bazel
// invocation. All executable paths are supplied by the proxy's declared inputs.
func runAction(args []string, stdin io.Reader, stdout, stderr io.Writer) (code int) {
	flags := flag.NewFlagSet("recording-sandbox --action", flag.ContinueOnError)
	flags.SetOutput(stderr)
	tracee := flags.String("tracee", "", "declared Tracee executable")
	outputPath := flags.String("trace-output", "", "declared recording output")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *tracee == "" || *outputPath == "" || flags.NArg() == 0 {
		fmt.Fprintln(stderr, "action mode requires --tracee, --trace-output, and a command after --")
		return 2
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	output, err := os.OpenFile(*outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer output.Close()
	stderr = &lockedWriter{Writer: stderr}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	trace, address, installDir, err := startTracee(*tracee, output, stderr)
	if err != nil {
		fmt.Fprintln(stderr, "starting Tracee:", err)
		return 1
	}
	defer os.RemoveAll(installDir)
	defer func() {
		if err := trace.stop(); err != nil && !trace.failureReported {
			fmt.Fprintln(stderr, "stopping Tracee:", err)
			if code == 0 {
				code = 1
			}
		}
	}()
	if err := waitForTracee(trace, address, signals); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	command := exec.Command(flags.Arg(0), flags.Args()[1:]...)
	command.Stdin, command.Stdout, command.Stderr = stdin, stdout, stderr
	action, err := startChild(command)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitCode(err)
	}
	for {
		select {
		case sig := <-signals:
			_ = action.cmd.Process.Signal(sig)
		case <-trace.done:
			_ = action.stop()
			fmt.Fprintln(stderr, "Tracee exited during action execution")
			trace.failureReported = true
			return 1
		case <-action.done:
			return exitCode(action.err)
		}
	}
}
