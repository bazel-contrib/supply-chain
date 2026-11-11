package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return path
}

func TestRun_ValidInstance(t *testing.T) {
	dir := t.TempDir()
	schema := writeFile(t, dir, "schema.json", `{
		"$id": "mem://schema.json",
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["name"],
		"properties": {"name": {"type": "string"}}
	}`)
	instance := writeFile(t, dir, "instance.json", `{"name": "ok"}`)
	output := filepath.Join(dir, "output")

	if err := run(schema, nil, instance, output); err != nil {
		t.Fatalf("run() = %v, want nil", err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output file to be written: %v", err)
	}
}

func TestRun_InvalidInstance(t *testing.T) {
	dir := t.TempDir()
	schema := writeFile(t, dir, "schema.json", `{
		"$id": "mem://schema.json",
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["name"],
		"properties": {"name": {"type": "string"}}
	}`)
	instance := writeFile(t, dir, "instance.json", `{"name": 123}`)
	output := filepath.Join(dir, "output")

	err := run(schema, nil, instance, output)
	if err == nil {
		t.Fatal("run() = nil, want a validation error")
	}
	if _, statErr := os.Stat(output); statErr == nil {
		t.Fatal("expected no output file to be written on validation failure")
	}
}

func TestRun_ResolvesAuxSchemaByItsOwnID(t *testing.T) {
	dir := t.TempDir()
	// The root schema's $ref deliberately differs from the aux file's path
	// on disk, to prove resolution goes through the registered "$id"
	// (http://example.test/aux.json) and not the filename.
	aux := writeFile(t, dir, "aux-on-disk.json", `{
		"$id": "http://example.test/aux.json",
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "string",
		"enum": ["red", "green", "blue"]
	}`)
	schema := writeFile(t, dir, "schema.json", `{
		"$id": "http://example.test/schema.json",
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {"color": {"$ref": "aux.json"}}
	}`)
	goodInstance := writeFile(t, dir, "good.json", `{"color": "red"}`)
	badInstance := writeFile(t, dir, "bad.json", `{"color": "purple"}`)

	if err := run(schema, []string{aux}, goodInstance, filepath.Join(dir, "out-good")); err != nil {
		t.Fatalf("run() with valid enum value = %v, want nil", err)
	}
	if err := run(schema, []string{aux}, badInstance, filepath.Join(dir, "out-bad")); err == nil {
		t.Fatal("run() with invalid enum value = nil, want an error")
	}
}

func TestRun_UnregisteredRefFailsWithoutNetworkAccess(t *testing.T) {
	dir := t.TempDir()
	// No http(s) loader is registered anywhere in this binary (we never
	// import the httploader extension), so a $ref to an unregistered http
	// URL must fail outright rather than attempt a network fetch.
	schema := writeFile(t, dir, "schema.json", `{
		"$id": "http://example.test/schema.json",
		"$schema": "http://json-schema.org/draft-07/schema#",
		"$ref": "http://example.test/not-vendored.json"
	}`)
	instance := writeFile(t, dir, "instance.json", `{}`)

	err := run(schema, nil, instance, filepath.Join(dir, "output"))
	if err == nil {
		t.Fatal("run() = nil, want an error for an unregistered $ref")
	}
}

func TestRun_SchemaMissingID(t *testing.T) {
	dir := t.TempDir()
	schema := writeFile(t, dir, "schema.json", `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object"
	}`)
	instance := writeFile(t, dir, "instance.json", `{}`)

	err := run(schema, nil, instance, filepath.Join(dir, "output"))
	if err == nil {
		t.Fatal("run() = nil, want an error for a schema without \"$id\"")
	}
}
