// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package remoteapisproxy

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"google.golang.org/protobuf/proto"
)

func TestActionLogCountsOriginalMetadata(t *testing.T) {
	p := newTestProxy(t, nil)
	var log bytes.Buffer
	p.config.ActionLog = &log
	put := func(message proto.Message) *repb.Digest {
		t.Helper()
		d, err := p.store.putProto(message)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	// Identical subtrees at distinct paths each contribute their file and link.
	leaf := put(&repb.Directory{
		Files:    []*repb.FileNode{{Name: "file", Digest: digest(nil)}},
		Symlinks: []*repb.SymlinkNode{{Name: "link", Target: "file"}},
	})
	root := put(&repb.Directory{
		Files:       []*repb.FileNode{{Name: "root-file", Digest: digest(nil)}},
		Directories: []*repb.DirectoryNode{{Name: "a", Digest: leaf}, {Name: "b", Digest: leaf}},
	})
	command := &repb.Command{
		Arguments:            []string{"echo", strings.Repeat("界", 65)},
		OutputPaths:          []string{"one", "two"},
		OutputFiles:          []string{"ignored-legacy-file"},
		OutputDirectories:    []string{"ignored-legacy-dir"},
		EnvironmentVariables: []*repb.Command_EnvironmentVariable{{Name: "A", Value: "1"}, {Name: "B", Value: "2"}},
		Platform:             &repb.Platform{Properties: []*repb.Platform_Property{{Name: "legacy", Value: "1"}}},
	}
	action := &repb.Action{CommandDigest: put(command), InputRootDigest: root,
		Platform: &repb.Platform{Properties: []*repb.Platform_Property{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}}},
	}
	if err := p.logAction(context.Background(), "instance", put(action)); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("action command=%q inputs=5 outputs=2 env=2 exec_properties=2\n", "echo "+strings.Repeat("界", 55))
	if got := log.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	// Deprecated output and platform fields still work when their replacements
	// are absent; embedded newlines must not create extra log lines.
	log.Reset()
	command.Arguments = []string{"echo", "one\ntwo"}
	command.OutputPaths = nil
	action.CommandDigest = put(command)
	action.Platform = nil
	if err := p.logAction(context.Background(), "instance", put(action)); err != nil {
		t.Fatal(err)
	}
	want = "action command=\"echo one\\ntwo\" inputs=5 outputs=2 env=2 exec_properties=1\n"
	if got := log.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActionLogExecutionAndCacheHit(t *testing.T) {
	p := newTestProxy(t, recordingRunner)
	var log bytes.Buffer
	p.config.ActionLog = &log
	conn := connect(t, p.Endpoint())
	ctx := clientContext(t, Route{})
	action := fixture(t, ctx, conn)
	execute(t, ctx, conn, action)
	execute(t, ctx, conn, action)
	p.logMu.Lock()
	defer p.logMu.Unlock()
	got := log.String()
	if strings.Count(got, "\n") != 1 || !strings.HasSuffix(got, "inputs=1 outputs=2 env=1 exec_properties=1\n") {
		t.Fatalf("expected one original-action summary, excluding tools/trace outputs and cache hits: %q", got)
	}
}

func TestActionLogDisabledAndConcurrentWrites(t *testing.T) {
	p := newTestProxy(t, nil)
	if err := p.logAction(context.Background(), "", nil); err != nil {
		t.Fatal("disabled logging read the action:", err)
	}
	root, _ := p.store.putProto(&repb.Directory{})
	command, _ := p.store.putProto(&repb.Command{Arguments: []string{"true"}})
	action, _ := p.store.putProto(&repb.Action{InputRootDigest: root, CommandDigest: command})
	var log bytes.Buffer
	p.config.ActionLog = &log
	var workers sync.WaitGroup
	for range 16 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := p.logAction(context.Background(), "", action); err != nil {
				t.Error(err)
			}
		}()
	}
	workers.Wait()
	want := strings.Repeat("action command=\"true\" inputs=0 outputs=0 env=0 exec_properties=0\n", 16)
	if got := log.String(); got != want {
		t.Fatalf("concurrent logs interleaved: %q", got)
	}
}
