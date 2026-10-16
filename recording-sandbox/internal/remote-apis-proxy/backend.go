// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package remoteapisproxy

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	bs "google.golang.org/genproto/googleapis/bytestream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type backend struct {
	conn     *grpc.ClientConn
	http     *http.Client
	base     string
	route    Route
	executor bool
}

func tlsConfig(route Route) (*tls.Config, error) {
	config := &tls.Config{MinVersion: tls.VersionTLS12}
	if route.Certificate != "" {
		data, err := os.ReadFile(route.Certificate)
		if err != nil {
			return nil, err
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("invalid upstream TLS certificate")
		}
		config.RootCAs = roots
	}
	if route.ClientCertificate != "" || route.ClientKey != "" {
		certificate, err := tls.LoadX509KeyPair(route.ClientCertificate, route.ClientKey)
		if err != nil {
			return nil, err
		}
		config.Certificates = []tls.Certificate{certificate}
	}
	return config, nil
}

func (p *Proxy) backend(endpoint string, route Route, executor bool) (*backend, error) {
	key := endpoint + "\x00" + route.Encode() + fmt.Sprint(executor)
	p.mu.Lock()
	defer p.mu.Unlock()
	if b := p.backends[key]; b != nil {
		return b, nil
	}
	tlsConf, err := tlsConfig(route)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		parsed, err := url.Parse(endpoint)
		if err != nil || parsed.Host == "" {
			return nil, status.Error(codes.InvalidArgument, "invalid cache URL")
		}
		b := &backend{route: route, executor: executor, base: strings.TrimRight(endpoint, "/"), http: &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConf}, Timeout: 2 * time.Minute}}
		p.backends[key] = b
		return b, nil
	}
	transport := credentials.NewTLS(tlsConf)
	target := strings.TrimPrefix(endpoint, "grpcs://")
	if strings.HasPrefix(endpoint, "grpc://") {
		target = strings.TrimPrefix(endpoint, "grpc://")
		transport = insecure.NewCredentials()
	}
	if target == "" || strings.Contains(target, "://") {
		return nil, status.Error(codes.InvalidArgument, "executor must use grpc:// or grpcs://")
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(transport), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16<<20)))
	if err != nil {
		return nil, err
	}
	b := &backend{conn: conn, route: route, executor: executor}
	p.backends[key] = b
	return b, nil
}

func (b *backend) context(ctx context.Context) context.Context {
	ctx = upstreamContext(ctx)
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	all := append(append(append([]string{}, b.route.Headers...), b.route.ExecutorHeaders...), b.route.CacheHeaders...)
	for _, header := range all {
		key, _, _ := strings.Cut(header, "=")
		md.Delete(strings.ToLower(key))
	}
	headers := append([]string{}, b.route.Headers...)
	if b.executor {
		headers = append(headers, b.route.ExecutorHeaders...)
	} else {
		headers = append(headers, b.route.CacheHeaders...)
	}
	for _, header := range headers {
		key, value, ok := strings.Cut(header, "=")
		if ok && !strings.EqualFold(key, RouteHeader) {
			md.Append(strings.ToLower(key), value)
		}
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func resource(instance string, d *repb.Digest) string {
	if instance != "" {
		instance += "/"
	}
	return instance + "blobs/" + digestKey(d)
}

func (b *backend) get(ctx context.Context, instance string, d *repb.Digest) ([]byte, error) {
	if b.http != nil {
		return b.httpRequest(ctx, "GET", "cas", d.Hash, nil)
	}
	stream, err := bs.NewByteStreamClient(b.conn).Read(b.context(ctx), &bs.ReadRequest{ResourceName: resource(instance, d)})
	if err != nil {
		return nil, err
	}
	var result bytes.Buffer
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			return result.Bytes(), nil
		}
		if err != nil {
			return nil, err
		}
		result.Write(response.Data)
	}
}

func (b *backend) put(ctx context.Context, instance string, d *repb.Digest, data []byte) error {
	if b.http != nil {
		_, err := b.httpRequest(ctx, "PUT", "cas", d.Hash, data)
		return err
	}
	stream, err := bs.NewByteStreamClient(b.conn).Write(b.context(ctx))
	if err != nil {
		return err
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return err
	}
	name := resource(instance, d)
	index := strings.LastIndex(name, "blobs/")
	name = name[:index] + "uploads/" + hex.EncodeToString(id[:]) + "/" + name[index:]
	for offset := 0; ; {
		end := min(offset+(64<<10), len(data))
		if err := stream.Send(&bs.WriteRequest{ResourceName: name, WriteOffset: int64(offset), Data: data[offset:end], FinishWrite: end == len(data)}); err != nil {
			return err
		}
		offset = end
		if end == len(data) {
			break
		}
	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	if response.CommittedSize != d.SizeBytes {
		return status.Error(codes.DataLoss, "upstream committed wrong blob size")
	}
	return nil
}

func (b *backend) getResult(ctx context.Context, instance string, d *repb.Digest) (*repb.ActionResult, error) {
	if b.http == nil {
		return repb.NewActionCacheClient(b.conn).GetActionResult(b.context(ctx), &repb.GetActionResultRequest{InstanceName: instance, ActionDigest: d})
	}
	data, err := b.httpRequest(ctx, "GET", "ac", d.Hash, nil)
	if err != nil {
		return nil, err
	}
	result := new(repb.ActionResult)
	err = proto.Unmarshal(data, result)
	return result, err
}

func (b *backend) putResult(ctx context.Context, instance string, d *repb.Digest, result *repb.ActionResult) error {
	if b.http == nil {
		_, err := repb.NewActionCacheClient(b.conn).UpdateActionResult(b.context(ctx), &repb.UpdateActionResultRequest{InstanceName: instance, ActionDigest: d, ActionResult: result})
		return err
	}
	data, err := proto.Marshal(result)
	if err != nil {
		return err
	}
	_, err = b.httpRequest(ctx, "PUT", "ac", d.Hash, data)
	return err
}

func (b *backend) httpRequest(ctx context.Context, method, kind, hash string, body []byte) ([]byte, error) {
	target, err := url.Parse(b.base)
	if err != nil {
		return nil, err
	}
	target.Path = strings.TrimRight(target.Path, "/") + "/" + kind + "/" + hash
	target.RawPath = ""
	request, err := http.NewRequestWithContext(ctx, method, target.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	md, _ := metadata.FromOutgoingContext(b.context(ctx))
	for key, values := range md {
		if strings.HasSuffix(key, "-bin") {
			continue
		}
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	if method == "PUT" {
		request.Header.Set("Content-Type", "application/octet-stream")
	}
	response, err := b.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil, status.Error(codes.NotFound, "upstream cache miss")
	}
	if response.StatusCode == http.StatusUnauthorized {
		return nil, status.Error(codes.Unauthenticated, "upstream cache authentication failed")
	}
	if response.StatusCode == http.StatusForbidden {
		return nil, status.Error(codes.PermissionDenied, "upstream cache access denied")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, status.Errorf(codes.Unavailable, "upstream cache returned HTTP %d", response.StatusCode)
	}
	return io.ReadAll(response.Body)
}
