// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

// Package remoteapisproxy implements an invocation-scoped Remote Execution API proxy.
package remoteapisproxy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RouteHeader carries the original upstreams after Bazel's endpoints are
// rewritten to the loopback proxy. It is consumed here, never forwarded.
const RouteHeader = "x-recording-sandbox-route"

type Route struct {
	Executor          string   `json:"executor,omitempty"`
	Cache             string   `json:"cache,omitempty"`
	Certificate       string   `json:"certificate,omitempty"`
	ClientCertificate string   `json:"client_certificate,omitempty"`
	ClientKey         string   `json:"client_key,omitempty"`
	Headers           []string `json:"headers,omitempty"`
	ExecutorHeaders   []string `json:"executor_headers,omitempty"`
	CacheHeaders      []string `json:"cache_headers,omitempty"`
}

func (r Route) Encode() string {
	data, _ := json.Marshal(r)
	return base64.RawURLEncoding.EncodeToString(data)
}

func routeFromContext(ctx context.Context) (Route, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get(RouteHeader)
	if len(values) == 0 {
		return Route{}, nil
	}
	if len(values) != 1 {
		return Route{}, status.Error(codes.InvalidArgument, "multiple proxy routing headers")
	}
	data, err := base64.RawURLEncoding.DecodeString(values[0])
	if err != nil {
		return Route{}, status.Error(codes.InvalidArgument, "invalid proxy routing header")
	}
	var route Route
	if err := json.Unmarshal(data, &route); err != nil {
		return route, status.Error(codes.InvalidArgument, "invalid proxy route")
	}
	return route, nil
}

func upstreamContext(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	md = md.Copy()
	for _, key := range []string{RouteHeader, ":authority", "content-type", "user-agent", "te", "grpc-timeout"} {
		md.Delete(key)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// RewriteFlags preserves user headers and instance names, replacing only the
// endpoints and our reserved routing headers. Canonicalized flags use '='.
func RewriteFlags(args []string, endpoint string) ([]string, Route, error) {
	var route Route
	var result []string
	var residue []string
	for i, arg := range args {
		if arg == "--" {
			residue = args[i:]
			args = args[:i]
			break
		}
	}
	for _, arg := range args {
		name, value, hasValue := strings.Cut(arg, "=")
		switch name {
		case "--remote_executor", "--remote_cache":
			if !hasValue {
				return nil, route, fmt.Errorf("%s requires an inline value", name)
			}
			if name == "--remote_executor" {
				route.Executor = value
			} else {
				route.Cache = value
			}
			continue
		case "--tls_certificate":
			route.Certificate = value
		case "--tls_client_certificate":
			route.ClientCertificate = value
		case "--tls_client_key":
			route.ClientKey = value
		case "--remote_header", "--remote_exec_header", "--remote_cache_header":
			key, _, _ := strings.Cut(value, "=")
			if strings.EqualFold(key, RouteHeader) {
				continue
			}
			switch name {
			case "--remote_header":
				route.Headers = append(route.Headers, value)
			case "--remote_exec_header":
				route.ExecutorHeaders = append(route.ExecutorHeaders, value)
			case "--remote_cache_header":
				route.CacheHeaders = append(route.CacheHeaders, value)
			}
		}
		result = append(result, arg)
	}
	result = append(result,
		"--remote_executor="+endpoint,
		"--remote_cache="+endpoint,
		"--remote_exec_header="+RouteHeader+"="+route.Encode(),
		"--remote_cache_header="+RouteHeader+"="+route.Encode(),
		"--remote_local_fallback=false",
	)
	return append(result, residue...), route, nil
}
