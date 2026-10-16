// Copyright 2026 The Bazel Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0.

package remoteapisproxy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type store struct{ root string }

func digest(data []byte) *repb.Digest {
	hash := sha256.Sum256(data)
	return &repb.Digest{Hash: hex.EncodeToString(hash[:]), SizeBytes: int64(len(data))}
}

func validDigest(d *repb.Digest) error {
	if d == nil || len(d.Hash) != 64 || d.SizeBytes < 0 || strings.ToLower(d.Hash) != d.Hash {
		return status.Error(codes.InvalidArgument, "invalid SHA-256 digest")
	}
	if _, err := hex.DecodeString(d.Hash); err != nil {
		return status.Error(codes.InvalidArgument, "invalid digest hash")
	}
	return nil
}

func digestKey(d *repb.Digest) string { return fmt.Sprintf("%s/%d", d.GetHash(), d.GetSizeBytes()) }

func (s store) get(d *repb.Digest) ([]byte, error) {
	if err := validDigest(d); err != nil {
		return nil, err
	}
	if d.SizeBytes == 0 && d.Hash == digest(nil).Hash {
		return nil, nil
	}
	data, err := os.ReadFile(filepath.Join(s.root, d.Hash))
	if os.IsNotExist(err) {
		return nil, status.Error(codes.NotFound, "blob not found: "+digestKey(d))
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !proto.Equal(d, digest(data)) {
		return nil, status.Error(codes.DataLoss, "CAS digest mismatch")
	}
	return data, nil
}

func (s store) put(d *repb.Digest, data []byte) error {
	if err := validDigest(d); err != nil {
		return err
	}
	if !proto.Equal(d, digest(data)) {
		return status.Error(codes.InvalidArgument, "uploaded blob does not match digest")
	}
	file, err := os.CreateTemp(s.root, "upload-*")
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
	return os.Rename(file.Name(), filepath.Join(s.root, d.Hash))
}

func (s store) putProto(message proto.Message) (*repb.Digest, error) {
	data, err := (proto.MarshalOptions{Deterministic: true}).Marshal(message)
	if err != nil {
		return nil, err
	}
	d := digest(data)
	return d, s.put(d, data)
}
