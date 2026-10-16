// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package remoteapisproxy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// safePath resolves existing parents, rejecting symlink escapes as well as
// lexical traversal. The last component is left for Lstat when collecting links.
func safePath(root, name string) (string, error) {
	if err := relativePath(name, true); err != nil {
		return "", err
	}
	target := filepath.Join(root, filepath.FromSlash(name))
	parent := filepath.Dir(target)
	if name == "" {
		return root, nil
	}
	for {
		resolved, err := filepath.EvalSymlinks(parent)
		if err == nil {
			rel, err := filepath.Rel(root, resolved)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return "", status.Error(codes.InvalidArgument, "action path escapes execution root")
			}
			return target, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		next := filepath.Dir(parent)
		if next == parent {
			return "", err
		}
		parent = next
	}
}

func safeLink(root, link, target string) error {
	if filepath.IsAbs(target) || strings.Contains(target, "\\") || strings.ContainsRune(target, 0) {
		return status.Error(codes.InvalidArgument, "absolute or invalid symlink target")
	}
	resolved := filepath.Clean(filepath.Join(filepath.Dir(link), target))
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return status.Error(codes.InvalidArgument, "symlink escapes execution root")
	}
	return nil
}

type pendingLink struct{ name, target string }

func (p *Proxy) materialize(ctx context.Context, instance, root string, d *repb.Digest) error {
	var links []pendingLink
	var visit func(string, *repb.Digest, int) error
	visit = func(dirPath string, d *repb.Digest, depth int) error {
		if depth > 256 {
			return status.Error(codes.InvalidArgument, "input tree is too deep")
		}
		dir := new(repb.Directory)
		if err := p.proto(ctx, instance, d, dir); err != nil {
			return err
		}
		names := map[string]bool{}
		check := func(name string) error {
			if strings.Contains(name, "/") || name == "" || name == "." || name == ".." || strings.Contains(name, "\\") || strings.ContainsRune(name, 0) || names[name] {
				return status.Error(codes.InvalidArgument, "invalid or duplicate input node name")
			}
			names[name] = true
			return nil
		}
		for _, file := range dir.Files {
			if err := check(file.Name); err != nil {
				return err
			}
			data, err := p.blob(ctx, instance, file.Digest)
			if err != nil {
				return err
			}
			mode := os.FileMode(0644)
			if file.IsExecutable {
				mode = 0755
			}
			if err := os.WriteFile(filepath.Join(dirPath, file.Name), data, mode); err != nil {
				return err
			}
		}
		for _, child := range dir.Directories {
			if err := check(child.Name); err != nil {
				return err
			}
			childPath := filepath.Join(dirPath, child.Name)
			if err := os.Mkdir(childPath, 0755); err != nil {
				return err
			}
			if err := visit(childPath, child.Digest, depth+1); err != nil {
				return err
			}
		}
		for _, link := range dir.Symlinks {
			if err := check(link.Name); err != nil {
				return err
			}
			linkPath := filepath.Join(dirPath, link.Name)
			if err := safeLink(root, linkPath, link.Target); err != nil {
				return err
			}
			links = append(links, pendingLink{linkPath, link.Target})
		}
		return nil
	}
	if err := visit(root, d, 0); err != nil {
		return err
	}
	for _, link := range links {
		if err := os.Symlink(link.target, link.name); err != nil {
			return err
		}
	}
	return nil
}

func (p *Proxy) executeLocal(ctx context.Context, instance string, action *repb.Action, command *repb.Command) (*repb.ActionResult, error) {
	root, err := os.MkdirTemp("", "recording-sandbox-action-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(root)
	if err := p.materialize(ctx, instance, root, action.InputRootDigest); err != nil {
		return nil, err
	}
	work, err := safePath(root, command.WorkingDirectory)
	if err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(work)
	if err != nil {
		return nil, err
	}
	if _, err := safePath(root, filepath.ToSlash(filepath.Join(command.WorkingDirectory, "placeholder"))); err != nil {
		return nil, err
	}
	work = resolved
	info, err := os.Stat(work)
	if err != nil || !info.IsDir() {
		return nil, status.Error(codes.InvalidArgument, "working directory does not exist")
	}
	outputs := append(append(append([]string{}, command.OutputPaths...), command.OutputFiles...), command.OutputDirectories...)
	for _, name := range outputs {
		if err := relativePath(name, false); err != nil {
			return nil, err
		}
		absolute, err := safePath(root, filepath.ToSlash(filepath.Join(command.WorkingDirectory, name)))
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Dir(absolute), 0755); err != nil {
			return nil, err
		}
	}
	var result *repb.ActionResult
	if p.config.Runner != nil {
		result, err = p.config.Runner(ctx, root, command)
	} else {
		result, err = runCommand(ctx, work, command)
	}
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "action runner returned no result")
	}
	for _, name := range outputs {
		absolute, err := safePath(root, filepath.ToSlash(filepath.Join(command.WorkingDirectory, name)))
		if err != nil {
			return nil, err
		}
		info, err := os.Lstat(absolute)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(absolute)
			if err != nil {
				return nil, err
			}
			if err := safeLink(root, absolute, target); err != nil {
				return nil, err
			}
			result.OutputSymlinks = append(result.OutputSymlinks, &repb.OutputSymlink{Path: name, Target: target})
		case info.IsDir():
			var children []*repb.Directory
			dir, d, err := p.collectDirectory(root, absolute, &children)
			if err != nil {
				return nil, err
			}
			unique := map[string]*repb.Directory{}
			for _, child := range children {
				childDigest, err := p.store.putProto(child)
				if err != nil {
					return nil, err
				}
				unique[childDigest.Hash] = child
			}
			keys := make([]string, 0, len(unique))
			for key := range unique {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			children = nil
			for _, key := range keys {
				children = append(children, unique[key])
			}
			treeDigest, err := p.store.putProto(&repb.Tree{Root: dir, Children: children})
			if err != nil {
				return nil, err
			}
			result.OutputDirectories = append(result.OutputDirectories, &repb.OutputDirectory{Path: name, TreeDigest: treeDigest, RootDirectoryDigest: d})
		case info.Mode().IsRegular():
			data, err := os.ReadFile(absolute)
			if err != nil {
				return nil, err
			}
			d := digest(data)
			if err := p.store.put(d, data); err != nil {
				return nil, err
			}
			result.OutputFiles = append(result.OutputFiles, &repb.OutputFile{Path: name, Digest: d, IsExecutable: info.Mode()&0111 != 0})
		default:
			return nil, status.Error(codes.InvalidArgument, "unsupported output file type")
		}
	}
	for _, item := range []struct {
		data []byte
		dest **repb.Digest
	}{{result.StdoutRaw, &result.StdoutDigest}, {result.StderrRaw, &result.StderrDigest}} {
		d := digest(item.data)
		if err := p.store.put(d, item.data); err != nil {
			return nil, err
		}
		*item.dest = d
	}
	if len(result.StdoutRaw) > 32<<10 {
		result.StdoutRaw = nil
	}
	if len(result.StderrRaw) > 32<<10 {
		result.StderrRaw = nil
	}
	return result, nil
}

func runCommand(ctx context.Context, work string, command *repb.Command) (*repb.ActionResult, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
	}
	name := command.Arguments[0]
	if !filepath.IsAbs(name) {
		name = filepath.Join(work, name)
	}
	cmd := exec.CommandContext(ctx, name, command.Arguments[1:]...)
	cmd.Dir = work
	cmd.Env = []string{}
	for _, variable := range command.EnvironmentVariables {
		if variable.Name == "" || strings.ContainsAny(variable.Name, "=\x00") || strings.ContainsRune(variable.Value, 0) {
			return nil, status.Error(codes.InvalidArgument, "invalid environment variable")
		}
		cmd.Env = append(cmd.Env, variable.Name+"="+variable.Value)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM) }
	cmd.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if cmd.Process != nil {
		defer syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	if ctx.Err() != nil {
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil, status.FromContextError(ctx.Err()).Err()
	}
	code := int32(0)
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			code = int32(exit.ExitCode())
			if code < 0 {
				code = 128 + int32(exit.Sys().(syscall.WaitStatus).Signal())
			}
		} else {
			return nil, status.Error(codes.Internal, fmt.Sprint(err))
		}
	}
	return &repb.ActionResult{ExitCode: code, StdoutRaw: stdout.Bytes(), StderrRaw: stderr.Bytes()}, nil
}

func (p *Proxy) collectDirectory(root, absolute string, children *[]*repb.Directory) (*repb.Directory, *repb.Digest, error) {
	entries, err := os.ReadDir(absolute)
	if err != nil {
		return nil, nil, err
	}
	dir := new(repb.Directory)
	for _, entry := range entries {
		full := filepath.Join(absolute, entry.Name())
		info, err := os.Lstat(full)
		if err != nil {
			return nil, nil, err
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(full)
			if err != nil {
				return nil, nil, err
			}
			if err := safeLink(root, full, target); err != nil {
				return nil, nil, err
			}
			dir.Symlinks = append(dir.Symlinks, &repb.SymlinkNode{Name: entry.Name(), Target: target})
		case info.IsDir():
			child, d, err := p.collectDirectory(root, full, children)
			if err != nil {
				return nil, nil, err
			}
			*children = append(*children, child)
			dir.Directories = append(dir.Directories, &repb.DirectoryNode{Name: entry.Name(), Digest: d})
		case info.Mode().IsRegular():
			data, err := os.ReadFile(full)
			if err != nil {
				return nil, nil, err
			}
			d := digest(data)
			if err := p.store.put(d, data); err != nil {
				return nil, nil, err
			}
			dir.Files = append(dir.Files, &repb.FileNode{Name: entry.Name(), Digest: d, IsExecutable: info.Mode()&0111 != 0})
		default:
			return nil, nil, status.Error(codes.InvalidArgument, "unsupported output node")
		}
	}
	sort.Slice(dir.Files, func(i, j int) bool { return dir.Files[i].Name < dir.Files[j].Name })
	sort.Slice(dir.Directories, func(i, j int) bool { return dir.Directories[i].Name < dir.Directories[j].Name })
	sort.Slice(dir.Symlinks, func(i, j int) bool { return dir.Symlinks[i].Name < dir.Symlinks[j].Name })
	d, err := p.store.putProto(dir)
	return dir, d, err
}
