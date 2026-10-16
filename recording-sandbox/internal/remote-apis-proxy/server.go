// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package remoteapisproxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	lropb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	semver "github.com/bazelbuild/remote-apis/build/bazel/semver"
	bs "google.golang.org/genproto/googleapis/bytestream"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Config struct {
	SandboxPath    string
	TraceePath     string
	TraceDirectory string
	// ActionLog enables execution summaries when non-nil.
	ActionLog io.Writer
	// Runner is an injectable action launcher; nil uses the packaged sandbox.
	Runner func(context.Context, string, *repb.Command) (*repb.ActionResult, error)
}

type Proxy struct {
	repb.UnimplementedExecutionServer
	repb.UnimplementedActionCacheServer
	repb.UnimplementedContentAddressableStorageServer
	repb.UnimplementedCapabilitiesServer
	bs.UnimplementedByteStreamServer
	lropb.UnimplementedOperationsServer
	config     Config
	store      store
	server     *grpc.Server
	listener   net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	logMu      sync.Mutex
	backends   map[string]*backend
	results    map[string]*repb.ActionResult
	operations map[string]*operation
	workers    sync.WaitGroup
	tools      *repb.Digest
}

func Start(config Config) (*Proxy, error) {
	root, err := os.MkdirTemp("", "recording-sandbox-proxy-*")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &Proxy{config: config, store: store{root}, ctx: ctx, cancel: cancel, backends: map[string]*backend{}, results: map[string]*repb.ActionResult{}, operations: map[string]*operation{}}
	fail := func(err error) (*Proxy, error) { cancel(); os.RemoveAll(root); return nil, err }
	if config.TraceDirectory != "" {
		if err := os.MkdirAll(config.TraceDirectory, 0700); err != nil {
			return fail(err)
		}
	}
	if err := p.prepareTools(); err != nil {
		return fail(err)
	}
	p.listener, err = net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return fail(err)
	}
	p.server = grpc.NewServer(grpc.MaxRecvMsgSize(16<<20), grpc.MaxSendMsgSize(16<<20), grpc.WaitForHandlers(true))
	repb.RegisterExecutionServer(p.server, p)
	repb.RegisterActionCacheServer(p.server, p)
	repb.RegisterContentAddressableStorageServer(p.server, p)
	repb.RegisterCapabilitiesServer(p.server, p)
	bs.RegisterByteStreamServer(p.server, p)
	lropb.RegisterOperationsServer(p.server, p)
	go p.server.Serve(p.listener)
	return p, nil
}

func (p *Proxy) Endpoint() string { return "grpc://" + p.listener.Addr().String() }
func (p *Proxy) Close() {
	p.cancel()
	p.server.Stop()
	p.workers.Wait()
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, b := range p.backends {
		if b.conn != nil {
			b.conn.Close()
		}
		if b.http != nil {
			b.http.CloseIdleConnections()
		}
	}
	os.RemoveAll(p.store.root)
}

func (p *Proxy) blob(ctx context.Context, instance string, d *repb.Digest) ([]byte, error) {
	data, err := p.store.get(d)
	if err == nil {
		return data, nil
	}
	if status.Code(err) != codes.NotFound {
		return nil, err
	}
	route, err := routeFromContext(ctx)
	if err != nil {
		return nil, err
	}
	for index, endpoint := range []string{route.Cache, route.Executor} {
		if endpoint == "" {
			continue
		}
		b, err := p.backend(endpoint, route, index == 1)
		if err != nil {
			return nil, err
		}
		data, err = b.get(ctx, instance, d)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				continue
			}
			return nil, err
		}
		if err := p.store.put(d, data); err != nil {
			return nil, status.Error(codes.DataLoss, err.Error())
		}
		return data, nil
	}
	return nil, status.Error(codes.NotFound, "blob not found: "+digestKey(d))
}

func (p *Proxy) proto(ctx context.Context, instance string, d *repb.Digest, message proto.Message) error {
	data, err := p.blob(ctx, instance, d)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(data, message); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}

func (p *Proxy) upload(ctx context.Context, instance string, d *repb.Digest, data []byte) error {
	if err := p.store.put(d, data); err != nil {
		return err
	}
	route, err := routeFromContext(ctx)
	if err != nil {
		return err
	}
	if route.Cache != "" {
		b, err := p.backend(route.Cache, route, false)
		if err != nil {
			return err
		}
		return b.put(ctx, instance, d, data)
	}
	return nil
}

func (p *Proxy) GetCapabilities(context.Context, *repb.GetCapabilitiesRequest) (*repb.ServerCapabilities, error) {
	return &repb.ServerCapabilities{
		LowApiVersion: &semver.SemVer{Major: 2}, HighApiVersion: &semver.SemVer{Major: 2, Minor: 3},
		CacheCapabilities:     &repb.CacheCapabilities{DigestFunctions: []repb.DigestFunction_Value{repb.DigestFunction_SHA256}, ActionCacheUpdateCapabilities: &repb.ActionCacheUpdateCapabilities{UpdateEnabled: true}, MaxBatchTotalSizeBytes: 4 << 20, SymlinkAbsolutePathStrategy: repb.SymlinkAbsolutePathStrategy_DISALLOWED},
		ExecutionCapabilities: &repb.ExecutionCapabilities{DigestFunction: repb.DigestFunction_SHA256, DigestFunctions: []repb.DigestFunction_Value{repb.DigestFunction_SHA256}, ExecEnabled: true},
	}, nil
}

func (p *Proxy) FindMissingBlobs(ctx context.Context, req *repb.FindMissingBlobsRequest) (*repb.FindMissingBlobsResponse, error) {
	response := new(repb.FindMissingBlobsResponse)
	for _, d := range req.BlobDigests {
		_, err := p.blob(ctx, req.InstanceName, d)
		if status.Code(err) == codes.NotFound {
			response.MissingBlobDigests = append(response.MissingBlobDigests, d)
		} else if err != nil {
			return nil, err
		}
	}
	return response, nil
}

func (p *Proxy) BatchUpdateBlobs(ctx context.Context, req *repb.BatchUpdateBlobsRequest) (*repb.BatchUpdateBlobsResponse, error) {
	response := new(repb.BatchUpdateBlobsResponse)
	for _, item := range req.Requests {
		var err error
		if item.Compressor != repb.Compressor_IDENTITY {
			err = status.Error(codes.InvalidArgument, "only identity compression is supported")
		} else {
			err = p.upload(ctx, req.InstanceName, item.Digest, item.Data)
		}
		response.Responses = append(response.Responses, &repb.BatchUpdateBlobsResponse_Response{Digest: item.Digest, Status: status.Convert(err).Proto()})
	}
	return response, nil
}

func (p *Proxy) BatchReadBlobs(ctx context.Context, req *repb.BatchReadBlobsRequest) (*repb.BatchReadBlobsResponse, error) {
	response := new(repb.BatchReadBlobsResponse)
	for _, d := range req.Digests {
		data, err := p.blob(ctx, req.InstanceName, d)
		response.Responses = append(response.Responses, &repb.BatchReadBlobsResponse_Response{Digest: d, Data: data, Status: status.Convert(err).Proto()})
	}
	return response, nil
}

func (p *Proxy) GetTree(req *repb.GetTreeRequest, stream repb.ContentAddressableStorage_GetTreeServer) error {
	queue := []*repb.Digest{req.RootDigest}
	seen := map[string]bool{}
	var dirs []*repb.Directory
	for len(queue) > 0 {
		d := queue[0]
		queue = queue[1:]
		if seen[digestKey(d)] {
			continue
		}
		seen[digestKey(d)] = true
		dir := new(repb.Directory)
		if err := p.proto(stream.Context(), req.InstanceName, d, dir); err != nil {
			return err
		}
		dirs = append(dirs, dir)
		for _, child := range dir.Directories {
			queue = append(queue, child.Digest)
		}
	}
	start := 0
	if req.PageToken != "" {
		n, err := strconv.Atoi(req.PageToken)
		if err != nil || n < 0 || n > len(dirs) {
			return status.Error(codes.InvalidArgument, "invalid page token")
		}
		start = n
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 100
	}
	for start < len(dirs) {
		end := min(start+pageSize, len(dirs))
		token := ""
		if end < len(dirs) {
			token = strconv.Itoa(end)
		}
		if err := stream.Send(&repb.GetTreeResponse{Directories: dirs[start:end], NextPageToken: token}); err != nil {
			return err
		}
		start = end
	}
	return nil
}

func parseResource(name string, upload bool) (string, *repb.Digest, error) {
	parts := strings.Split(name, "/")
	index := -1
	for i, part := range parts {
		if part == "blobs" && i+2 == len(parts)-1 {
			index = i
			break
		}
	}
	if index < 0 {
		return "", nil, status.Error(codes.InvalidArgument, "invalid ByteStream resource")
	}
	size, err := strconv.ParseInt(parts[index+2], 10, 64)
	if err != nil {
		return "", nil, status.Error(codes.InvalidArgument, "invalid resource size")
	}
	d := &repb.Digest{Hash: parts[index+1], SizeBytes: size}
	if err := validDigest(d); err != nil {
		return "", nil, err
	}
	instance := parts[:index]
	if upload {
		if len(instance) < 2 || instance[len(instance)-2] != "uploads" || instance[len(instance)-1] == "" {
			return "", nil, status.Error(codes.InvalidArgument, "invalid upload resource")
		}
		instance = instance[:len(instance)-2]
	}
	return strings.Join(instance, "/"), d, nil
}

func (p *Proxy) Read(req *bs.ReadRequest, stream bs.ByteStream_ReadServer) error {
	instance, d, err := parseResource(req.ResourceName, false)
	if err != nil {
		return err
	}
	data, err := p.blob(stream.Context(), instance, d)
	if err != nil {
		return err
	}
	if req.ReadOffset < 0 || req.ReadOffset > int64(len(data)) || req.ReadLimit < 0 {
		return status.Error(codes.OutOfRange, "invalid read range")
	}
	data = data[req.ReadOffset:]
	if req.ReadLimit > 0 && req.ReadLimit < int64(len(data)) {
		data = data[:req.ReadLimit]
	}
	for len(data) > 0 {
		n := min(64<<10, len(data))
		if err := stream.Send(&bs.ReadResponse{Data: data[:n]}); err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

func (p *Proxy) Write(stream bs.ByteStream_WriteServer) error {
	var data []byte
	var name, instance string
	var d *repb.Digest
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return status.Error(codes.InvalidArgument, "missing finish_write")
		}
		if err != nil {
			return err
		}
		if name == "" {
			name = req.ResourceName
			instance, d, err = parseResource(name, true)
			if err != nil {
				return err
			}
		} else if req.ResourceName != "" && req.ResourceName != name {
			return status.Error(codes.InvalidArgument, "resource changed within upload")
		}
		if req.WriteOffset != int64(len(data)) {
			return status.Error(codes.OutOfRange, "uploads must start at offset zero and be contiguous")
		}
		if int64(len(data))+int64(len(req.Data)) > d.SizeBytes {
			return status.Error(codes.OutOfRange, "upload exceeds declared size")
		}
		data = append(data, req.Data...)
		if req.FinishWrite {
			if err := p.upload(stream.Context(), instance, d, data); err != nil {
				return err
			}
			return stream.SendAndClose(&bs.WriteResponse{CommittedSize: int64(len(data))})
		}
	}
}

func (p *Proxy) QueryWriteStatus(ctx context.Context, req *bs.QueryWriteStatusRequest) (*bs.QueryWriteStatusResponse, error) {
	_, d, err := parseResource(req.ResourceName, true)
	if err != nil {
		return nil, err
	}
	_, err = p.store.get(d)
	if err != nil {
		return nil, err
	}
	return &bs.QueryWriteStatusResponse{CommittedSize: d.SizeBytes, Complete: true}, nil
}

func resultKey(ctx context.Context, instance string, d *repb.Digest) (string, error) {
	if err := validDigest(d); err != nil {
		return "", err
	}
	route, err := routeFromContext(ctx)
	if err != nil {
		return "", err
	}
	return route.Encode() + "\x00" + instance + "\x00" + digestKey(d), nil
}

func (p *Proxy) GetActionResult(ctx context.Context, req *repb.GetActionResultRequest) (*repb.ActionResult, error) {
	key, err := resultKey(ctx, req.InstanceName, req.ActionDigest)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	cached := p.results[key]
	p.mu.Unlock()
	if cached != nil {
		return proto.Clone(cached).(*repb.ActionResult), nil
	}
	route, _ := routeFromContext(ctx)
	if route.Cache != "" {
		b, err := p.backend(route.Cache, route, false)
		if err != nil {
			return nil, err
		}
		result, err := b.getResult(ctx, req.InstanceName, req.ActionDigest)
		if err != nil {
			return nil, err
		}
		if err := p.retainTrace(ctx, req.InstanceName, req.ActionDigest, result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, status.Error(codes.NotFound, "action cache miss")
}

func (p *Proxy) UpdateActionResult(ctx context.Context, req *repb.UpdateActionResultRequest) (*repb.ActionResult, error) {
	key, err := resultKey(ctx, req.InstanceName, req.ActionDigest)
	if err != nil {
		return nil, err
	}
	if req.ActionResult == nil {
		return nil, status.Error(codes.InvalidArgument, "missing action result")
	}
	route, _ := routeFromContext(ctx)
	if route.Cache != "" {
		b, err := p.backend(route.Cache, route, false)
		if err != nil {
			return nil, err
		}
		if err := b.putResult(ctx, req.InstanceName, req.ActionDigest, req.ActionResult); err != nil {
			return nil, err
		}
	}
	p.mu.Lock()
	p.results[key] = proto.Clone(req.ActionResult).(*repb.ActionResult)
	p.mu.Unlock()
	return req.ActionResult, nil
}

func (p *Proxy) GetOperation(ctx context.Context, req *lropb.GetOperationRequest) (*lropb.Operation, error) {
	op, err := p.getOperation(req.Name)
	if err != nil {
		return nil, err
	}
	op.mu.Lock()
	defer op.mu.Unlock()
	return proto.Clone(op.latest).(*lropb.Operation), nil
}
func (p *Proxy) CancelOperation(ctx context.Context, req *lropb.CancelOperationRequest) (*emptypb.Empty, error) {
	op, err := p.getOperation(req.Name)
	if err != nil {
		return nil, err
	}
	op.cancel()
	return &emptypb.Empty{}, nil
}

func rpcStatus(err error) *statuspb.Status {
	if status.Code(err) == codes.Unknown {
		return status.New(codes.Internal, fmt.Sprint(err)).Proto()
	}
	return status.Convert(err).Proto()
}
func traceFile(directory string, d *repb.Digest) string {
	return filepath.Join(directory, d.Hash+".jsonl")
}
