package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func portOf(t *testing.T, label string) string {
	t.Helper()
	var ports map[string]string
	if err := json.Unmarshal([]byte(os.Getenv("ASSIGNED_PORTS")), &ports); err != nil {
		t.Fatalf("reading ASSIGNED_PORTS: %v", err)
	}
	port, ok := ports[label]
	if !ok {
		t.Fatalf("no port assigned to %s in %v", label, ports)
	}
	return port
}

func get(t *testing.T, url string) (int, string) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading %s: %v", url, err)
	}
	return resp.StatusCode, string(body)
}

func TestServer(t *testing.T) {
	base := "http://127.0.0.1:" + portOf(t, "@@//:server_service")
	for path, want := range map[string]string{
		"/":             "<html",
		"/frontend.js":  "",
		"/openapi.json": "openapi",
		"/api/python":   `"service": "python"`,
		"/api/java":     `"service":"java"`,
	} {
		status, body := get(t, base+path)
		if status != http.StatusOK {
			t.Errorf("GET %s: status %d, body %q", path, status, body)
		}
		if !strings.Contains(body, want) {
			t.Errorf("GET %s: body %q does not contain %q", path, body, want)
		}
	}
}

func TestNodeService(t *testing.T) {
	_, body := get(t, "http://127.0.0.1:"+portOf(t, "@@//:node_service_itest"))
	if !strings.Contains(body, "Node.js runtime") {
		t.Errorf("unexpected body %q", body)
	}
}
