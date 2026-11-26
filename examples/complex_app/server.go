package main

import (
	"embed"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed frontend.js web/index.html
var frontend embed.FS

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/frontend.js" {
			http.ServeFileFS(w, r, frontend, "frontend.js")
			return
		}
		http.ServeFileFS(w, r, frontend, "web/index.html")
	})
	http.HandleFunc("/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, openAPISpec)
	})
	pythonAPI := flag.String("python-api", "/app/python_api", "Path of the Python API binary.")
	javaGreeter := flag.String("java-greeter", "/app/java-greeter", "Path of the Java greeter binary.")
	portFlag := flag.String("port", "", "Port to listen on, defaults to $PORT or 8080.")
	flag.Parse()

	http.HandleFunc("/api/python", commandHandler(*pythonAPI))
	http.HandleFunc("/api/java", commandHandler(*javaGreeter))

	port := *portFlag
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// resolveBinary maps a Bazel runfiles path, as passed by the integration test,
// to a file; plain paths such as /app/python_api are returned as is.
func resolveBinary(path string) string {
	if dir := os.Getenv("RUNFILES_DIR"); dir != "" && !filepath.IsAbs(path) {
		return filepath.Join(dir, path)
	}
	return path
}

func commandHandler(path string) http.HandlerFunc {
	path = resolveBinary(path)
	return func(w http.ResponseWriter, _ *http.Request) {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			http.Error(w, "service is not included in this libc variant", http.StatusNotImplemented)
			return
		}
		output, err := exec.Command(path).CombinedOutput()
		if err != nil {
			http.Error(w, string(output), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(output)
	}
}
