// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/bazel-contrib/supply-chain/recording-sandbox/internal/remote-apis-proxy"
)

func startProxy(traceePath, traceDirectory string, actionLog io.Writer) (*remoteapisproxy.Proxy, error) {
	sandbox, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return remoteapisproxy.Start(remoteapisproxy.Config{SandboxPath: sandbox, TraceePath: traceePath, TraceDirectory: traceDirectory, ActionLog: actionLog})
}

func runProxy(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("recording-sandbox --proxy", flag.ContinueOnError)
	flags.SetOutput(stderr)
	traceDirectory := flags.String("trace-directory", "recording-sandbox-actions", "directory for action recordings")
	logActions := flags.Bool("log-actions", false, "log command prefixes and action input/output/environment/platform counts to stderr")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		return 2
	}
	tracee, err := bundledTracee()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	var actionLog io.Writer
	if *logActions {
		actionLog = stderr
	}
	proxy, err := startProxy(tracee, *traceDirectory, actionLog)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer proxy.Close()
	fmt.Fprintln(stdout, proxy.Endpoint())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	<-signals
	return 0
}
