// Command schemavalidate validates a JSON document against a JSON Schema,
// entirely offline.
//
// It is meant to back a Bazel validation action (see
// https://bazel.build/extending/rules#validation-actions): the schema and
// any auxiliary schemas it references via "$ref" must be supplied as inputs
// (--schema / --aux_schema), and on success the tool writes --output, which
// the calling rule declares as its validation action's output.
//
// This package deliberately does not import
// github.com/santhosh-tekuri/jsonschema/v5/httploader (or any other loader
// extension), so the only registered URL scheme is "file". A schema that
// references a URL not supplied via --schema/--aux_schema fails the build
// with a clear "no Loader found" error instead of silently reaching out to
// the network.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func main() {
	var schemaPath, instancePath, outputPath string
	var auxPaths stringList

	flag.StringVar(&schemaPath, "schema", "", "Path to the root JSON Schema document.")
	flag.StringVar(&instancePath, "instance", "", "Path to the JSON document to validate.")
	flag.StringVar(&outputPath, "output", "", "Path to write on successful validation.")
	flag.Var(&auxPaths, "aux_schema", "Path to an additional schema referenced by --schema via \"$ref\" (repeatable). Registered under its own \"$id\".")
	flag.Parse()

	if schemaPath == "" || instancePath == "" || outputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --schema, --instance and --output are required")
		os.Exit(1)
	}

	if err := run(schemaPath, auxPaths, instancePath, outputPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(schemaPath string, auxPaths []string, instancePath, outputPath string) error {
	compiler := jsonschema.NewCompiler()

	rootID, err := addSchemaResource(compiler, schemaPath)
	if err != nil {
		return fmt.Errorf("loading schema %s: %w", schemaPath, err)
	}

	for _, auxPath := range auxPaths {
		if _, err := addSchemaResource(compiler, auxPath); err != nil {
			return fmt.Errorf("loading auxiliary schema %s: %w", auxPath, err)
		}
	}

	schema, err := compiler.Compile(rootID)
	if err != nil {
		return fmt.Errorf("compiling schema %s: %w", schemaPath, err)
	}

	instance, err := decodeJSON(instancePath)
	if err != nil {
		return fmt.Errorf("reading instance %s: %w", instancePath, err)
	}

	if err := schema.Validate(instance); err != nil {
		return fmt.Errorf("%s does not conform to %s:\n\n%v", instancePath, schemaPath, err)
	}

	if err := os.WriteFile(outputPath, []byte("OK\n"), 0o644); err != nil {
		return fmt.Errorf("writing output %s: %w", outputPath, err)
	}
	return nil
}

// addSchemaResource reads a vendored schema file and registers it with the
// compiler under its own "$id", so that a "$ref" in another document which
// resolves to this same "$id" is served from this already-loaded resource
// instead of being fetched.
func addSchemaResource(compiler *jsonschema.Compiler, path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var meta struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return "", err
	}
	if meta.ID == "" {
		return "", fmt.Errorf("schema has no \"$id\"")
	}

	if err := compiler.AddResource(meta.ID, bytes.NewReader(data)); err != nil {
		return "", err
	}
	return meta.ID, nil
}

// decodeJSON decodes like encoding/json, but with UseNumber() so that
// schema.Validate sees the same number representation the jsonschema
// package itself uses when it loads a document, avoiding float64-precision
// mismatches on "type": "integer" checks.
func decodeJSON(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var v interface{}
	if err := decoder.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

type stringList []string

func (s *stringList) String() string {
	return fmt.Sprint([]string(*s))
}

func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}
