package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadBuildTimestamp_NoPath(t *testing.T) {
	got, err := readBuildTimestamp("")
	if err != nil {
		t.Fatalf("readBuildTimestamp(\"\") error = %v", err)
	}
	if !got.Equal(time.Unix(0, 0).UTC()) {
		t.Errorf("readBuildTimestamp(\"\") = %v, want the Unix epoch", got)
	}
}

func TestReadBuildTimestamp_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "volatile-status.txt")
	if err := os.WriteFile(path, []byte("BUILD_SCM_REVISION abc123\nBUILD_TIMESTAMP 1700000000\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := readBuildTimestamp(path)
	if err != nil {
		t.Fatalf("readBuildTimestamp() error = %v", err)
	}
	want := time.Unix(1700000000, 0).UTC()
	if !got.Equal(want) {
		t.Errorf("readBuildTimestamp() = %v, want %v", got, want)
	}
}

func TestReadBuildTimestamp_MissingKeyFallsBackToEpoch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "volatile-status.txt")
	if err := os.WriteFile(path, []byte("BUILD_SCM_REVISION abc123\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := readBuildTimestamp(path)
	if err != nil {
		t.Fatalf("readBuildTimestamp() error = %v", err)
	}
	if !got.Equal(time.Unix(0, 0).UTC()) {
		t.Errorf("readBuildTimestamp() = %v, want the Unix epoch", got)
	}
}
