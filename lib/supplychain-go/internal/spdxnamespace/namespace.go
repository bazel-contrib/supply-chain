// Package spdxnamespace reads and validates SPDX document namespaces.
package spdxnamespace

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// ReadFile reads a document namespace from the file at path and validates it.
// An empty path returns an empty namespace, meaning "not provided".
func ReadFile(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	namespace := strings.TrimSpace(string(data))
	if err := Validate(namespace); err != nil {
		return "", fmt.Errorf("document namespace file %s: %w", path, err)
	}
	return namespace, nil
}

// Validate checks the constraints SPDX puts on a document namespace: it must be
// an absolute URI and must not contain a fragment.
func Validate(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace must not be empty")
	}
	uri, err := url.Parse(namespace)
	if err != nil {
		return err
	}
	if !uri.IsAbs() {
		return fmt.Errorf("namespace must be an absolute URI with a scheme")
	}
	if uri.Fragment != "" || strings.Contains(namespace, "#") {
		return fmt.Errorf("namespace must not contain a fragment ('#')")
	}
	return nil
}
