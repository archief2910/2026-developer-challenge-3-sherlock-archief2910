package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sherlock/internal/analysis"
	"sherlock/internal/block"
	"sherlock/internal/report"
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
	mux.HandleFunc("/api/upload", handleUpload)

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

func handleUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		writeJSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "PARSE_ERROR", "Failed to parse form: "+err.Error())
		return
	}

	blkFile, blkHeader, err := r.FormFile("blk")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "MISSING_FILE", "Missing blk.dat file")
		return
	}
	defer blkFile.Close()

	revFile, _, err := r.FormFile("rev")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "MISSING_FILE", "Missing rev.dat file")
		return
	}
	defer revFile.Close()

	xorFile, _, err := r.FormFile("xor")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "MISSING_FILE", "Missing xor.dat file")
		return
	}
	defer xorFile.Close()

	tempDir, err := os.MkdirTemp("", "sherlock-upload-*")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "TEMP_ERROR", "Failed to create temp directory")
		return
	}
	defer os.RemoveAll(tempDir)

	blkPath := filepath.Join(tempDir, blkHeader.Filename)
	revPath := filepath.Join(tempDir, "rev.dat")
	xorPath := filepath.Join(tempDir, "xor.dat")

	blkOut, err := os.Create(blkPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "WRITE_ERROR", "Failed to create blk file")
		return
	}
	if _, err := io.Copy(blkOut, blkFile); err != nil {
		blkOut.Close()
		writeJSONError(w, http.StatusInternalServerError, "WRITE_ERROR", "Failed to write blk file")
		return
	}
	blkOut.Close()

	revOut, err := os.Create(revPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "WRITE_ERROR", "Failed to create rev file")
		return
	}
	if _, err := io.Copy(revOut, revFile); err != nil {
		revOut.Close()
		writeJSONError(w, http.StatusInternalServerError, "WRITE_ERROR", "Failed to write rev file")
		return
	}
	revOut.Close()

	xorOut, err := os.Create(xorPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "WRITE_ERROR", "Failed to create xor file")
		return
	}
	if _, err := io.Copy(xorOut, xorFile); err != nil {
		xorOut.Close()
		writeJSONError(w, http.StatusInternalServerError, "WRITE_ERROR", "Failed to write xor file")
		return
	}
	xorOut.Close()

	parsedBlocks, err := block.ProcessBlockFiles(blkPath, revPath, xorPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "PARSE_ERROR", "Failed to parse block files: "+err.Error())
		return
	}

	blkFilename := blkHeader.Filename
	result := analysis.AnalyzeBlocks(parsedBlocks, blkFilename)

	stem := analysis.GetBlockStem(blkFilename)

	os.MkdirAll("out", 0755)

	jsonPath := filepath.Join("out", stem+".json")
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "JSON_ERROR", "Failed to marshal JSON: "+err.Error())
		return
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "WRITE_ERROR", "Failed to write JSON: "+err.Error())
		return
	}

	mdPath := filepath.Join("out", stem+".md")
	mdContent := report.GenerateMarkdownReport(result)
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "WRITE_ERROR", "Failed to write report: "+err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":     true,
		"file":   stem,
		"result": result,
	})
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
