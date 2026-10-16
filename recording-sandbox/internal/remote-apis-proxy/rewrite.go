// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package remoteapisproxy

import (
	"context"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"

	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const ToolsDirectory = ".recording-sandbox-tools"
const TraceOutput = ".recording-sandbox/trace.jsonl"

func (p *Proxy) prepareTools() error {
	dir := new(repb.Directory)
	for _, file := range []struct{ name, path string }{{"recording-sandbox", p.config.SandboxPath}, {"tracee-static", p.config.TraceePath}} {
		data, err := os.ReadFile(file.path)
		if err != nil {
			return err
		}
		d := digest(data)
		if err := p.store.put(d, data); err != nil {
			return err
		}
		dir.Files = append(dir.Files, &repb.FileNode{Name: file.name, Digest: d, IsExecutable: true})
	}
	d, err := p.store.putProto(dir)
	p.tools = d
	return err
}

func (p *Proxy) rewrite(ctx context.Context, instance string, original *repb.Digest) (*repb.Digest, *repb.Action, *repb.Command, error) {
	action := new(repb.Action)
	if err := p.proto(ctx, instance, original, action); err != nil {
		return nil, nil, nil, err
	}
	command := new(repb.Command)
	if err := p.proto(ctx, instance, action.CommandDigest, command); err != nil {
		return nil, nil, nil, err
	}
	if len(command.Arguments) == 0 {
		return nil, nil, nil, status.Error(codes.InvalidArgument, "empty action command")
	}
	if err := relativePath(command.WorkingDirectory, true); err != nil {
		return nil, nil, nil, err
	}
	platform := action.Platform
	if platform == nil {
		platform = command.Platform
	}
	for _, property := range platform.GetProperties() {
		if property.Name == "OSFamily" && !strings.EqualFold(property.Value, "linux") {
			return nil, nil, nil, status.Error(codes.FailedPrecondition, "recording sandbox requires Linux workers")
		}
		if property.Name == "ISA" || property.Name == "Arch" {
			v := strings.ToLower(property.Value)
			if v == "x86_64" {
				v = "amd64"
			}
			if v == "aarch64" {
				v = "arm64"
			}
			if v != runtime.GOARCH {
				return nil, nil, nil, status.Error(codes.FailedPrecondition, "worker architecture differs from bundled sandbox")
			}
		}
	}
	root := new(repb.Directory)
	if err := p.proto(ctx, instance, action.InputRootDigest, root); err != nil {
		return nil, nil, nil, err
	}
	for _, file := range root.Files {
		if file.Name == ToolsDirectory {
			return nil, nil, nil, status.Error(codes.InvalidArgument, "reserved sandbox input path already exists")
		}
	}
	for _, dir := range root.Directories {
		if dir.Name == ToolsDirectory {
			return nil, nil, nil, status.Error(codes.InvalidArgument, "reserved sandbox input path already exists")
		}
	}
	for _, link := range root.Symlinks {
		if link.Name == ToolsDirectory {
			return nil, nil, nil, status.Error(codes.InvalidArgument, "reserved sandbox input path already exists")
		}
	}
	root.Directories = append(root.Directories, &repb.DirectoryNode{Name: ToolsDirectory, Digest: p.tools})
	sort.Slice(root.Directories, func(i, j int) bool { return root.Directories[i].Name < root.Directories[j].Name })
	var err error
	action.InputRootDigest, err = p.store.putProto(root)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, output := range append(append(append([]string{}, command.OutputPaths...), command.OutputFiles...), command.OutputDirectories...) {
		if err := relativePath(output, false); err != nil {
			return nil, nil, nil, err
		}
		if output == TraceOutput || strings.HasPrefix(output, ".recording-sandbox/") || output == ".recording-sandbox" {
			return nil, nil, nil, status.Error(codes.InvalidArgument, "reserved sandbox output path already declared")
		}
	}
	prefix := ""
	if command.WorkingDirectory != "" {
		prefix = strings.Repeat("../", len(strings.Split(command.WorkingDirectory, "/")))
	}
	arguments := []string{prefix + ToolsDirectory + "/recording-sandbox", "--action", "--tracee=" + prefix + ToolsDirectory + "/tracee-static", "--trace-output=" + TraceOutput, "--"}
	command.Arguments = append(arguments, command.Arguments...)
	if len(command.OutputPaths) > 0 {
		command.OutputPaths = append(command.OutputPaths, TraceOutput)
		sort.Strings(command.OutputPaths)
	} else {
		command.OutputFiles = append(command.OutputFiles, TraceOutput)
		sort.Strings(command.OutputFiles)
	}
	action.CommandDigest, err = p.store.putProto(command)
	if err != nil {
		return nil, nil, nil, err
	}
	d, err := p.store.putProto(action)
	return d, action, command, err
}

func relativePath(name string, empty bool) error {
	if empty && name == "" {
		return nil
	}
	if name == "" || strings.Contains(name, "\\") || strings.ContainsRune(name, 0) || path.IsAbs(name) || path.Clean(name) != name || name == "." || name == ".." || strings.HasPrefix(name, "../") {
		return status.Errorf(codes.InvalidArgument, "invalid relative action path %q", name)
	}
	return nil
}

func (p *Proxy) syncBlob(ctx context.Context, instance string, b *backend, d *repb.Digest) error {
	data, err := p.blob(ctx, instance, d)
	if err != nil {
		return err
	}
	return b.put(ctx, instance, d, data)
}

func (p *Proxy) syncInputs(ctx context.Context, instance string, b *backend, d *repb.Digest, seen map[string]bool) error {
	if seen[digestKey(d)] {
		return nil
	}
	seen[digestKey(d)] = true
	dir := new(repb.Directory)
	if err := p.proto(ctx, instance, d, dir); err != nil {
		return err
	}
	for _, file := range dir.Files {
		if err := p.syncBlob(ctx, instance, b, file.Digest); err != nil {
			return err
		}
	}
	for _, child := range dir.Directories {
		if err := p.syncInputs(ctx, instance, b, child.Digest, seen); err != nil {
			return err
		}
	}
	return p.syncBlob(ctx, instance, b, d)
}
