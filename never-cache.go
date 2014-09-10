package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func maybeSetContentTypeFromPath(w http.ResponseWriter, path string) {
	extensionMimeMap := map[string]string{
		".css": "text/css",
		".js":  "text/javascript",
	}
	mimeType, ok := extensionMimeMap[filepath.Ext(path)]
	if ok {
		w.Header().Add("Content-Type", mimeType)
	}
	return
}

func checkPath(path string, root string) (string, error) {
	if len(path) == 0 {
		path = "index.html"
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if len(absPath) <= len(root) || absPath[0:len(root)] != root || !os.IsPathSeparator(absPath[len(root)]) {
		return "", fmt.Errorf("\"%s\" is not within \"%s\"", path, root)
	}
	return absPath, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
        cwd, err := os.Getwd()
        var path string
	if err == nil {
		path, err = checkPath(r.URL.Path[1:], cwd)
	}
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, "Not Found", 404)
		return
	}
	log.Print(r.URL.Path)
	w.Header().Add("Cache-Control", "private, max-age=0, no-cache")
	http.ServeFile(w, r, path)
}

func main() {
        var port int
	flag.IntVar(&port, "port", 8080, "Serve files under the current directory on this port")
	flag.Parse()

	http.HandleFunc("/", handler)
	log.Print("Starting...")
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
