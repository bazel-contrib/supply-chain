package main

import (
	"embed"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	http.HandleFunc("/api/python", commandHandler("/app/python_api"))
	http.HandleFunc("/api/java", commandHandler("/app/java-greeter"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func commandHandler(path string) http.HandlerFunc {
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
