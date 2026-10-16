// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package remoteapisproxy

import (
	"context"
	"fmt"
	"strings"

	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (p *Proxy) logAction(ctx context.Context, instance string, original *repb.Digest) error {
	if p.config.ActionLog == nil {
		return nil
	}
	action := new(repb.Action)
	if err := p.proto(ctx, instance, original, action); err != nil {
		return err
	}
	command := new(repb.Command)
	if err := p.proto(ctx, instance, action.CommandDigest, command); err != nil {
		return err
	}
	inputs, err := p.countInputs(ctx, instance, action.InputRootDigest, 0)
	if err != nil {
		return err
	}
	outputs := len(command.OutputPaths)
	if outputs == 0 {
		outputs = len(command.OutputFiles) + len(command.OutputDirectories)
	}
	platform := action.Platform
	if platform == nil {
		platform = command.Platform
	}
	prefix := []rune(strings.Join(command.Arguments, " "))
	if len(prefix) > 60 {
		prefix = prefix[:60]
	}
	// Serialize whole lines and quote control characters so concurrent actions
	// and commands containing newlines still produce one summary per line.
	p.logMu.Lock()
	defer p.logMu.Unlock()
	fmt.Fprintf(p.config.ActionLog, "action command=%q inputs=%d outputs=%d env=%d exec_properties=%d\n",
		string(prefix), inputs, outputs, len(command.EnvironmentVariables), len(platform.GetProperties()))
	return nil
}

// Count file and symlink paths, excluding directory containers and injected
// sandbox inputs. A shared subtree counts once for each path it appears at.
func (p *Proxy) countInputs(ctx context.Context, instance string, digest *repb.Digest, depth int) (int, error) {
	if depth > 256 {
		return 0, status.Error(codes.InvalidArgument, "input tree is too deep")
	}
	directory := new(repb.Directory)
	if err := p.proto(ctx, instance, digest, directory); err != nil {
		return 0, err
	}
	count := len(directory.Files) + len(directory.Symlinks)
	for _, child := range directory.Directories {
		inputs, err := p.countInputs(ctx, instance, child.Digest, depth+1)
		if err != nil {
			return 0, err
		}
		count += inputs
	}
	return count, nil
}
