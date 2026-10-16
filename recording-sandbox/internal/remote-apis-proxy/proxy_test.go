// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package remoteapisproxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	lropb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	bs "google.golang.org/genproto/googleapis/bytestream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

func newTestProxy(t *testing.T, runner func(context.Context, string, *repb.Command) (*repb.ActionResult, error)) *Proxy {
	t.Helper()
	dir := t.TempDir()
	tool := filepath.Join(dir, "tool")
	if err := os.WriteFile(tool, []byte("pinned executable payload"), 0700); err != nil {
		t.Fatal(err)
	}
	p, err := Start(Config{SandboxPath: tool, TraceePath: tool, TraceDirectory: filepath.Join(dir, "traces"), Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

func connect(t *testing.T, endpoint string) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient(strings.TrimPrefix(endpoint, "grpc://"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func clientContext(t *testing.T, route Route) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(RouteHeader, route.Encode(), "authorization", "test-secret"))
}

func upload(t *testing.T, ctx context.Context, conn *grpc.ClientConn, message proto.Message) *repb.Digest {
	t.Helper()
	data, err := proto.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	return uploadBytes(t, ctx, conn, data)
}

func uploadBytes(t *testing.T, ctx context.Context, conn *grpc.ClientConn, data []byte) *repb.Digest {
	t.Helper()
	d := digest(data)
	response, err := repb.NewContentAddressableStorageClient(conn).BatchUpdateBlobs(ctx, &repb.BatchUpdateBlobsRequest{InstanceName: "instance", Requests: []*repb.BatchUpdateBlobsRequest_Request{{Digest: d, Data: data}}})
	if err != nil || len(response.GetResponses()) != 1 || response.Responses[0].Status.GetCode() != 0 {
		t.Fatalf("upload: %v %v", response, err)
	}
	return d
}

func fixture(t *testing.T, ctx context.Context, conn *grpc.ClientConn) *repb.Digest {
	t.Helper()
	file := uploadBytes(t, ctx, conn, []byte("hello"))
	empty := upload(t, ctx, conn, &repb.Directory{})
	work := upload(t, ctx, conn, &repb.Directory{Files: []*repb.FileNode{{Name: "input", Digest: file}}, Directories: []*repb.DirectoryNode{{Name: "directory", Digest: empty}}})
	root := upload(t, ctx, conn, &repb.Directory{Directories: []*repb.DirectoryNode{{Name: "work", Digest: work}}})
	command := upload(t, ctx, conn, &repb.Command{Arguments: []string{"/bin/sh", "-c", "IFS= read -r line < input; printf %s \"$line\" > output; printf nested > directory/file; printf stdout; printf stderr >&2"}, EnvironmentVariables: []*repb.Command_EnvironmentVariable{{Name: "PATH", Value: "/usr/bin:/bin"}}, WorkingDirectory: "work", OutputPaths: []string{"directory", "output"}})
	return upload(t, ctx, conn, &repb.Action{CommandDigest: command, InputRootDigest: root, Timeout: durationpb.New(10 * time.Second), Platform: &repb.Platform{Properties: []*repb.Platform_Property{{Name: "OSFamily", Value: "Linux"}}}})
}

// The tracer itself is tested by the executable's action-mode tests. Here we
// verify the REAPI transport and run the original command while simulating its
// recording output, without requiring kernel eBPF privileges in unit tests.
func recordingRunner(ctx context.Context, root string, command *repb.Command) (*repb.ActionResult, error) {
	if len(command.Arguments) < 7 || command.Arguments[1] != "--action" {
		return nil, fmt.Errorf("command was not wrapped: %q", command.Arguments)
	}
	for _, name := range []string{"recording-sandbox", "tracee-static"} {
		info, err := os.Stat(filepath.Join(root, ToolsDirectory, name))
		if err != nil || info.Mode()&0111 == 0 {
			return nil, fmt.Errorf("missing executable input %s", name)
		}
	}
	work := filepath.Join(root, command.WorkingDirectory)
	if err := os.MkdirAll(filepath.Join(work, ".recording-sandbox"), 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(work, TraceOutput), []byte("{\"event\":\"execve\"}\n"), 0600); err != nil {
		return nil, err
	}
	original := proto.Clone(command).(*repb.Command)
	original.Arguments = original.Arguments[5:]
	return runCommand(ctx, work, original)
}

func execute(t *testing.T, ctx context.Context, conn *grpc.ClientConn, d *repb.Digest) (*repb.ExecuteResponse, string) {
	t.Helper()
	stream, err := repb.NewExecutionClient(conn).Execute(ctx, &repb.ExecuteRequest{InstanceName: "instance", ActionDigest: d})
	if err != nil {
		t.Fatal(err)
	}
	name := ""
	for {
		op, err := stream.Recv()
		if err != nil {
			t.Fatal(err)
		}
		if name != "" && name != op.Name {
			t.Fatal("operation name changed")
		}
		name = op.Name
		if !op.Done {
			continue
		}
		if op.GetError() != nil {
			t.Fatalf("execution: %v", op.GetError())
		}
		response := new(repb.ExecuteResponse)
		if err := op.GetResponse().UnmarshalTo(response); err != nil {
			t.Fatal(err)
		}
		return response, name
	}
}

func checkResult(t *testing.T, p *Proxy, ctx context.Context, conn *grpc.ClientConn, d *repb.Digest, response *repb.ExecuteResponse) {
	t.Helper()
	if response.Result == nil || response.Result.ExitCode != 0 {
		t.Fatalf("result: %v", response)
	}
	readCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(RouteHeader, Route{}.Encode()))
	want := map[string]string{"output": "hello", TraceOutput: "{\"event\":\"execve\"}\n"}
	for _, file := range response.Result.OutputFiles {
		if expected, ok := want[file.Path]; ok {
			data, err := p.blob(readCtx, "instance", file.Digest)
			if err != nil || string(data) != expected {
				t.Fatalf("output %s: %s %v", file.Path, data, err)
			}
			delete(want, file.Path)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing outputs: %v", want)
	}
	if len(response.Result.OutputDirectories) != 1 {
		t.Fatal("missing directory output")
	}
	tree := new(repb.Tree)
	if err := p.proto(readCtx, "instance", response.Result.OutputDirectories[0].TreeDigest, tree); err != nil {
		t.Fatal(err)
	}
	if len(tree.Root.Files) != 1 || tree.Root.Files[0].Name != "file" {
		t.Fatalf("output tree: %v", tree)
	}
	if string(response.Result.StdoutRaw) != "stdout" || string(response.Result.StderrRaw) != "stderr" {
		t.Fatalf("stdio: %v", response.Result)
	}
	if _, err := os.Stat(traceFile(p.config.TraceDirectory, d)); err != nil {
		t.Fatal(err)
	}
	cached, err := repb.NewActionCacheClient(conn).GetActionResult(ctx, &repb.GetActionResultRequest{InstanceName: "instance", ActionDigest: d})
	if err != nil || !proto.Equal(cached, response.Result) {
		t.Fatalf("cache: %v %v", cached, err)
	}
}

func TestLocalExecutionAndWait(t *testing.T) {
	p := newTestProxy(t, recordingRunner)
	conn := connect(t, p.Endpoint())
	ctx := clientContext(t, Route{})
	d := fixture(t, ctx, conn)
	response, name := execute(t, ctx, conn, d)
	checkResult(t, p, ctx, conn, d, response)
	wait, err := repb.NewExecutionClient(conn).WaitExecution(ctx, &repb.WaitExecutionRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	op, err := wait.Recv()
	if err != nil || !op.Done || op.Name != name {
		t.Fatalf("wait: %v %v", op, err)
	}
	cached, _ := execute(t, ctx, conn, d)
	if !cached.CachedResult {
		t.Fatal("execution did not use action cache")
	}
}

type upstreamExecution struct {
	repb.UnimplementedExecutionServer
	storage  *Proxy
	original *repb.Digest
	called   atomic.Bool
}

func (s *upstreamExecution) Execute(req *repb.ExecuteRequest, stream repb.Execution_ExecuteServer) error {
	md, _ := metadata.FromIncomingContext(stream.Context())
	if len(md.Get(RouteHeader)) != 0 || len(md.Get("authorization")) != 1 || md.Get("authorization")[0] != "test-secret" {
		return status.Error(codes.PermissionDenied, "routing/auth headers were not handled correctly")
	}
	if len(md.Get("x-only-executor")) != 1 || md.Get("x-only-executor")[0] != "execute" || len(md.Get("x-only-cache")) != 0 {
		return status.Error(codes.PermissionDenied, "executor/cache credentials were mixed")
	}
	if proto.Equal(req.ActionDigest, s.original) {
		return status.Error(codes.InvalidArgument, "original action was forwarded without rewriting")
	}
	action := new(repb.Action)
	if err := s.storage.proto(stream.Context(), req.InstanceName, req.ActionDigest, action); err != nil {
		return err
	}
	command := new(repb.Command)
	if err := s.storage.proto(stream.Context(), req.InstanceName, action.CommandDigest, command); err != nil {
		return err
	}
	if command.Arguments[0] != "../"+ToolsDirectory+"/recording-sandbox" || !reflect.DeepEqual(command.OutputPaths, []string{TraceOutput, "directory", "output"}) || action.Timeout.AsDuration() != 10*time.Second || action.Platform.Properties[0].Value != "Linux" {
		return status.Error(codes.InvalidArgument, "rewritten action lost semantics")
	}
	s.called.Store(true)
	result, err := s.storage.executeLocal(stream.Context(), req.InstanceName, action, command)
	if err != nil {
		return err
	}
	encoded, _ := anypb.New(&repb.ExecuteResponse{Result: result})
	return stream.Send(&lropb.Operation{Name: "upstream-operation", Done: true, Result: &lropb.Operation_Response{Response: encoded}})
}

func TestRemoteExecutionAndSeparateCache(t *testing.T) {
	storage := newTestProxy(t, recordingRunner)
	execution := &upstreamExecution{storage: storage}
	server := grpc.NewServer()
	repb.RegisterExecutionServer(server, execution)
	repb.RegisterContentAddressableStorageServer(server, storage)
	bs.RegisterByteStreamServer(server, storage)
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)
	t.Cleanup(server.Stop)
	cache := newTestProxy(t, nil)
	route := Route{Executor: "grpc://" + listener.Addr().String(), Cache: cache.Endpoint(), ExecutorHeaders: []string{"x-only-executor=execute"}, CacheHeaders: []string{"x-only-cache=cache"}}
	p := newTestProxy(t, nil)
	conn := connect(t, p.Endpoint())
	ctx := clientContext(t, route)
	d := fixture(t, ctx, conn)
	execution.original = d
	response, _ := execute(t, ctx, conn, d)
	if !execution.called.Load() {
		t.Fatal("upstream executor not used")
	}
	checkResult(t, p, ctx, conn, d, response)
	// A second proxy can use the configured cache, including the recording and
	// nested directory blobs, without executing the original unwrapped action.
	p2 := newTestProxy(t, nil)
	conn2 := connect(t, p2.Endpoint())
	response2, _ := execute(t, ctx, conn2, d)
	if !response2.CachedResult {
		t.Fatal("separate remote cache was not used")
	}
	if _, err := os.Stat(traceFile(p2.config.TraceDirectory, d)); err != nil {
		t.Fatal("cached recording was not retained:", err)
	}
	for _, file := range response2.Result.OutputFiles {
		data, err := cache.store.get(file.Digest)
		if err != nil || len(data) == 0 {
			t.Fatalf("output not mirrored to cache: %v", err)
		}
	}
}

func TestHTTPRemoteCache(t *testing.T) {
	var mu sync.Mutex
	blobs := map[string][]byte{}
	var badHeader atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(RouteHeader) != "" || r.Header.Get("authorization") != "test-secret" || r.Header.Get("x-only-cache") != "cache" || r.Header.Get("x-only-executor") != "" {
			badHeader.Store(true)
		}
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case "PUT":
			data, _ := io.ReadAll(r.Body)
			blobs[r.URL.Path] = data
			w.WriteHeader(200)
		case "GET":
			data, ok := blobs[r.URL.Path]
			if !ok {
				w.WriteHeader(404)
				return
			}
			w.Write(data)
		default:
			w.WriteHeader(405)
		}
	}))
	defer server.Close()
	route := Route{Cache: server.URL + "/cache-prefix", CacheHeaders: []string{"x-only-cache=cache"}, ExecutorHeaders: []string{"x-only-executor=execute"}}
	p := newTestProxy(t, recordingRunner)
	conn := connect(t, p.Endpoint())
	ctx := clientContext(t, route)
	d := fixture(t, ctx, conn)
	response, _ := execute(t, ctx, conn, d)
	checkResult(t, p, ctx, conn, d, response)
	p2 := newTestProxy(t, nil)
	conn2 := connect(t, p2.Endpoint())
	cached, _ := execute(t, ctx, conn2, d)
	if !cached.CachedResult || badHeader.Load() {
		t.Fatal("HTTP cache routing/authentication failed")
	}
	if _, err := os.Stat(traceFile(p2.config.TraceDirectory, d)); err != nil {
		t.Fatal("HTTP-cached recording was not retained:", err)
	}
}

func TestByteStreamAndDigestValidation(t *testing.T) {
	p := newTestProxy(t, nil)
	conn := connect(t, p.Endpoint())
	ctx := clientContext(t, Route{})
	client := bs.NewByteStreamClient(conn)
	data := []byte(strings.Repeat("data", 20000))
	d := digest(data)
	name := "instance/uploads/test/blobs/" + digestKey(d)
	write, err := client.Write(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := write.Send(&bs.WriteRequest{ResourceName: name, Data: data[:30000]}); err != nil {
		t.Fatal(err)
	}
	if err := write.Send(&bs.WriteRequest{WriteOffset: 30000, Data: data[30000:], FinishWrite: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := write.CloseAndRecv(); err != nil {
		t.Fatal(err)
	}
	read, err := client.Read(ctx, &bs.ReadRequest{ResourceName: resource("instance", d), ReadOffset: 3, ReadLimit: 100})
	if err != nil {
		t.Fatal(err)
	}
	var got []byte
	for {
		value, err := read.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, value.Data...)
	}
	if !reflect.DeepEqual(got, data[3:103]) {
		t.Fatal("ByteStream read range changed data")
	}
	write, _ = client.Write(ctx)
	_ = write.Send(&bs.WriteRequest{ResourceName: name, WriteOffset: 1, Data: data, FinishWrite: true})
	if _, err := write.CloseAndRecv(); status.Code(err) != codes.OutOfRange {
		t.Fatalf("accepted non-contiguous upload: %v", err)
	}
	response, err := repb.NewContentAddressableStorageClient(conn).BatchUpdateBlobs(ctx, &repb.BatchUpdateBlobsRequest{Requests: []*repb.BatchUpdateBlobsRequest_Request{{Digest: d, Data: []byte("wrong")}}})
	if err != nil || codes.Code(response.Responses[0].Status.Code) != codes.InvalidArgument {
		t.Fatalf("accepted corrupt blob: %v %v", response, err)
	}
}

func TestCancellationAndReconnect(t *testing.T) {
	started := make(chan struct{})
	p := newTestProxy(t, func(ctx context.Context, root string, command *repb.Command) (*repb.ActionResult, error) {
		close(started)
		<-ctx.Done()
		return nil, status.FromContextError(ctx.Err()).Err()
	})
	conn := connect(t, p.Endpoint())
	ctx := clientContext(t, Route{})
	d := fixture(t, ctx, conn)
	streamCtx, disconnect := context.WithCancel(ctx)
	stream, err := repb.NewExecutionClient(conn).Execute(streamCtx, &repb.ExecuteRequest{InstanceName: "instance", ActionDigest: d})
	if err != nil {
		t.Fatal(err)
	}
	op, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	disconnect()
	<-started
	wait, err := repb.NewExecutionClient(conn).WaitExecution(ctx, &repb.WaitExecutionRequest{Name: op.Name})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lropb.NewOperationsClient(conn).CancelOperation(ctx, &lropb.CancelOperationRequest{Name: op.Name}); err != nil {
		t.Fatal(err)
	}
	for {
		done, err := wait.Recv()
		if err != nil {
			t.Fatal(err)
		}
		if done.Done {
			if codes.Code(done.GetError().GetCode()) != codes.Canceled {
				t.Fatalf("cancellation: %v", done)
			}
			break
		}
	}
}

func TestRewriteFlagsAndUnsafePaths(t *testing.T) {
	flags, route, err := RewriteFlags([]string{"--remote_executor=grpcs://executor", "--remote_cache=https://cache", "--remote_header=authorization=secret", "//:target", "--", "--remote_cache=program-argument"}, "grpc://127.0.0.1:1234")
	if err != nil {
		t.Fatal(err)
	}
	if route.Executor != "grpcs://executor" || route.Cache != "https://cache" || flags[len(flags)-1] != "--remote_cache=program-argument" {
		t.Fatalf("flags: %q route %v", flags, route)
	}
	for _, name := range []string{"../escape", "/absolute", "a/../../escape", "a\\b", "."} {
		if err := relativePath(name, false); err == nil {
			t.Errorf("accepted unsafe path %q", name)
		}
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	if _, err := safePath(root, "escape/file"); err == nil {
		t.Fatal("accepted output through escaping symlink")
	}
}
