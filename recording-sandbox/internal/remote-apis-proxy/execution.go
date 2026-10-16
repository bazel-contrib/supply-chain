// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package remoteapisproxy

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	lropb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type operation struct {
	mu      sync.Mutex
	latest  *lropb.Operation
	changed chan struct{}
	cancel  context.CancelFunc
}

func (op *operation) update(value *lropb.Operation) {
	op.mu.Lock()
	defer op.mu.Unlock()
	value.Name = op.latest.Name
	op.latest = value
	close(op.changed)
	op.changed = make(chan struct{})
}

func (p *Proxy) getOperation(name string) (*operation, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	op := p.operations[name]
	if op == nil {
		return nil, status.Error(codes.NotFound, "operation not found")
	}
	return op, nil
}

func (p *Proxy) Execute(req *repb.ExecuteRequest, stream repb.Execution_ExecuteServer) error {
	if err := validDigest(req.ActionDigest); err != nil {
		return err
	}
	if req.DigestFunction != repb.DigestFunction_UNKNOWN && req.DigestFunction != repb.DigestFunction_SHA256 {
		return status.Error(codes.InvalidArgument, "only SHA-256 is supported")
	}
	if _, err := routeFromContext(stream.Context()); err != nil {
		return err
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(p.ctx)
	md, _ := metadata.FromIncomingContext(stream.Context())
	ctx = metadata.NewIncomingContext(ctx, md.Copy())
	meta, _ := anypb.New(&repb.ExecuteOperationMetadata{Stage: repb.ExecutionStage_QUEUED, ActionDigest: req.ActionDigest})
	op := &operation{latest: &lropb.Operation{Name: "recording-sandbox/" + hex.EncodeToString(id[:]), Metadata: meta}, changed: make(chan struct{}), cancel: cancel}
	p.mu.Lock()
	p.operations[op.latest.Name] = op
	p.workers.Add(1)
	p.mu.Unlock()
	request := proto.Clone(req).(*repb.ExecuteRequest)
	go func() {
		defer p.workers.Done()
		defer cancel()
		response, err := p.execute(ctx, request, op)
		value := &lropb.Operation{Done: true}
		if err != nil {
			value.Result = &lropb.Operation_Error{Error: rpcStatus(err)}
		} else {
			encoded, _ := anypb.New(response)
			value.Result = &lropb.Operation_Response{Response: encoded}
		}
		completed, _ := anypb.New(&repb.ExecuteOperationMetadata{Stage: repb.ExecutionStage_COMPLETED, ActionDigest: req.ActionDigest})
		value.Metadata = completed
		op.update(value)
	}()
	return streamOperation(stream.Context(), op, stream.Send)
}

func (p *Proxy) WaitExecution(req *repb.WaitExecutionRequest, stream repb.Execution_WaitExecutionServer) error {
	op, err := p.getOperation(req.Name)
	if err != nil {
		return err
	}
	return streamOperation(stream.Context(), op, stream.Send)
}

func streamOperation(ctx context.Context, op *operation, send func(*lropb.Operation) error) error {
	for {
		op.mu.Lock()
		value := proto.Clone(op.latest).(*lropb.Operation)
		changed := op.changed
		op.mu.Unlock()
		if err := send(value); err != nil {
			return err
		}
		if value.Done {
			return nil
		}
		select {
		case <-ctx.Done():
			return status.FromContextError(ctx.Err()).Err()
		case <-changed:
		}
	}
}

func (p *Proxy) execute(ctx context.Context, req *repb.ExecuteRequest, op *operation) (*repb.ExecuteResponse, error) {
	if !req.SkipCacheLookup {
		result, err := p.GetActionResult(ctx, &repb.GetActionResultRequest{InstanceName: req.InstanceName, ActionDigest: req.ActionDigest})
		if err == nil {
			return &repb.ExecuteResponse{Result: result, CachedResult: true}, nil
		}
		if status.Code(err) != codes.NotFound {
			return nil, err
		}
	}
	modified, action, command, err := p.rewrite(ctx, req.InstanceName, req.ActionDigest)
	if err != nil {
		return nil, err
	}
	if action.Timeout != nil {
		if err := action.Timeout.CheckValid(); err != nil || action.Timeout.AsDuration() <= 0 {
			return nil, status.Error(codes.InvalidArgument, "invalid action timeout")
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, action.Timeout.AsDuration())
		defer cancel()
	}
	if err := p.logAction(ctx, req.InstanceName, req.ActionDigest); err != nil {
		return nil, err
	}
	meta, _ := anypb.New(&repb.ExecuteOperationMetadata{Stage: repb.ExecutionStage_EXECUTING, ActionDigest: req.ActionDigest})
	op.update(&lropb.Operation{Metadata: meta})
	route, _ := routeFromContext(ctx)
	var response *repb.ExecuteResponse
	if route.Executor == "" {
		result, err := p.executeLocal(ctx, req.InstanceName, action, command)
		if err != nil {
			return nil, err
		}
		response = &repb.ExecuteResponse{Result: result}
	} else {
		b, err := p.backend(route.Executor, route, true)
		if err != nil {
			return nil, err
		}
		if b.conn == nil {
			return nil, status.Error(codes.InvalidArgument, "executor must be a gRPC endpoint")
		}
		if err := p.syncInputs(ctx, req.InstanceName, b, action.InputRootDigest, map[string]bool{}); err != nil {
			return nil, err
		}
		for _, d := range []*repb.Digest{action.CommandDigest, modified} {
			if err := p.syncBlob(ctx, req.InstanceName, b, d); err != nil {
				return nil, err
			}
		}
		upstreamReq := proto.Clone(req).(*repb.ExecuteRequest)
		upstreamReq.ActionDigest = modified
		stream, err := repb.NewExecutionClient(b.conn).Execute(b.context(ctx), upstreamReq)
		if err != nil {
			return nil, err
		}
		name := ""
		defer func() {
			if ctx.Err() != nil && name != "" {
				cancelCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				md, _ := metadata.FromIncomingContext(ctx)
				cancelCtx = metadata.NewIncomingContext(cancelCtx, md)
				_, _ = lropb.NewOperationsClient(b.conn).CancelOperation(b.context(cancelCtx), &lropb.CancelOperationRequest{Name: name})
			}
		}()
		for {
			value, err := stream.Recv()
			if err == io.EOF {
				return nil, status.Error(codes.Unavailable, "upstream execution ended without a result")
			}
			if err != nil {
				return nil, err
			}
			name = value.Name
			if value.Done {
				if value.GetError() != nil {
					return nil, status.FromProto(value.GetError()).Err()
				}
				response = new(repb.ExecuteResponse)
				if value.GetResponse() == nil {
					return nil, status.Error(codes.DataLoss, "upstream returned no ExecuteResponse")
				}
				if err := value.GetResponse().UnmarshalTo(response); err != nil {
					return nil, err
				}
				break
			}
			// Keep the public operation name and original Action digest stable.
			copy := proto.Clone(value).(*lropb.Operation)
			if copy.Metadata != nil {
				m := new(repb.ExecuteOperationMetadata)
				if copy.Metadata.UnmarshalTo(m) == nil {
					m.ActionDigest = req.ActionDigest
					copy.Metadata, _ = anypb.New(m)
				}
			}
			op.update(copy)
		}
	}
	if response.Result == nil {
		if response.Status != nil && response.Status.Code != 0 {
			return response, nil
		}
		return nil, status.Error(codes.DataLoss, "executor returned no action result")
	}
	if err := p.persistResult(ctx, req.InstanceName, req.ActionDigest, response.Result); err != nil {
		return nil, err
	}
	if !action.DoNotCache && response.Result.ExitCode == 0 && (response.Status == nil || response.Status.Code == 0) {
		if _, err := p.UpdateActionResult(ctx, &repb.UpdateActionResultRequest{InstanceName: req.InstanceName, ActionDigest: req.ActionDigest, ActionResult: response.Result}); err != nil {
			return nil, err
		}
	}
	return response, nil
}

// Pull result blobs into the local CAS, mirror them to an optional cache, and
// retain recordings outside temporary execution roots before they disappear.
func (p *Proxy) persistResult(ctx context.Context, instance string, actionDigest *repb.Digest, result *repb.ActionResult) error {
	copyBlob := func(d *repb.Digest) error {
		if d == nil {
			return nil
		}
		data, err := p.blob(ctx, instance, d)
		if err != nil {
			return err
		}
		return p.upload(ctx, instance, d, data)
	}
	for _, d := range []*repb.Digest{result.StdoutDigest, result.StderrDigest} {
		if err := copyBlob(d); err != nil {
			return err
		}
	}
	for _, file := range result.OutputFiles {
		if len(file.Contents) > 0 {
			if err := p.store.put(file.Digest, file.Contents); err != nil {
				return err
			}
		}
		if err := copyBlob(file.Digest); err != nil {
			return err
		}
	}
	for _, output := range result.OutputDirectories {
		if output.TreeDigest != nil {
			tree := new(repb.Tree)
			if err := p.proto(ctx, instance, output.TreeDigest, tree); err != nil {
				return err
			}
			if err := copyBlob(output.TreeDigest); err != nil {
				return err
			}
			for _, dir := range append([]*repb.Directory{tree.Root}, tree.Children...) {
				if dir == nil {
					return status.Error(codes.DataLoss, "missing output Tree root")
				}
				for _, file := range dir.Files {
					if err := copyBlob(file.Digest); err != nil {
						return err
					}
				}
			}
		}
		if output.RootDirectoryDigest != nil {
			route, _ := routeFromContext(ctx)
			if route.Cache != "" {
				b, err := p.backend(route.Cache, route, false)
				if err != nil {
					return err
				}
				if err := p.syncInputs(ctx, instance, b, output.RootDirectoryDigest, map[string]bool{}); err != nil {
					return err
				}
			}
		}
	}
	return p.retainTrace(ctx, instance, actionDigest, result)
}

func (p *Proxy) retainTrace(ctx context.Context, instance string, actionDigest *repb.Digest, result *repb.ActionResult) error {
	if p.config.TraceDirectory == "" {
		return nil
	}
	for _, file := range result.OutputFiles {
		if file.Path == TraceOutput {
			data, err := p.blob(ctx, instance, file.Digest)
			if err != nil {
				return err
			}
			file, err := os.CreateTemp(p.config.TraceDirectory, ".trace-*")
			if err != nil {
				return err
			}
			defer os.Remove(file.Name())
			if _, err := file.Write(data); err != nil {
				file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
			return os.Rename(file.Name(), filepath.Clean(traceFile(p.config.TraceDirectory, actionDigest)))
		}
	}
	return nil
}
