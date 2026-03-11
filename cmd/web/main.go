package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/blocks", handleListBlocks)
	mux.HandleFunc("/api/blocks/", handleGetBlock)

	// Serve static frontend
	webDistPath := "web/dist"
	if _, err := os.Stat(webDistPath); err == nil {
		fileServer := http.FileServer(http.Dir(webDistPath))
		mux.Handle("/", spaHandler(fileServer, webDistPath))
	} else {
		// Fallback: serve a simple page
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Sherlock</title></head><body><h1>Sherlock Chain Analyzer</h1><p>Web UI not built. Run setup.sh first.</p></body></html>`)
		})
	}

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	fmt.Printf("http://%s\n", addr)

	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func handleListBlocks(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/blocks" {
		handleGetBlock(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	entries, err := os.ReadDir("out")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    true,
			"files": []string{},
		})
		return
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    true,
		"files": files,
	})
}

func handleGetBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract block name from URL: /api/blocks/<name>
	name := strings.TrimPrefix(r.URL.Path, "/api/blocks/")
	if name == "" {
		writeJSONError(w, http.StatusBadRequest, "MISSING_NAME", "Block name required")
		return
	}

	// Sanitize
	name = filepath.Base(name)
	if !strings.HasSuffix(name, ".json") {
		name += ".json"
	}

	filePath := filepath.Join("out", name)
	data, err := os.ReadFile(filePath)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Block file not found: %s", name))
		return
	}

	w.Write(data)
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// spaHandler handles SPA routing — serves index.html for unknown paths
type spaHandlerStruct struct {
	fileServer http.Handler
	staticDir  string
}

func spaHandler(fileServer http.Handler, staticDir string) http.Handler {
	return &spaHandlerStruct{fileServer: fileServer, staticDir: staticDir}
}

func (h *spaHandlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.staticDir, r.URL.Path)

	// Check if the file exists
	fi, err := os.Stat(path)
	if os.IsNotExist(err) || (err == nil && fi.IsDir()) {
		// Serve index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(h.staticDir, "index.html"))
		return
	}

	h.fileServer.ServeHTTP(w, r)
}
