package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"sherlock/internal/analysis"
	"sherlock/internal/block"
	"sherlock/internal/report"
)

func convertMSYSPath(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}
	re := regexp.MustCompile(`^/([a-zA-Z])/(.*)$`)
	if matches := re.FindStringSubmatch(path); matches != nil {
		drive := strings.ToUpper(matches[1])
		rest := strings.ReplaceAll(matches[2], "/", "\\")
		return drive + ":\\" + rest
	}
	return path
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "--block" {
		writeError("INVALID_ARGS", "Usage: sherlock-cli --block <blk.dat> <rev.dat> <xor.dat>")
		os.Exit(1)
	}

	if len(os.Args) < 5 {
		writeError("INVALID_ARGS", "Block mode requires: --block <blk.dat> <rev.dat> <xor.dat>")
		os.Exit(1)
	}

	blkPath := convertMSYSPath(os.Args[2])
	revPath := convertMSYSPath(os.Args[3])
	xorPath := convertMSYSPath(os.Args[4])

	// Validate files exist
	for _, f := range []string{blkPath, revPath, xorPath} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			writeError("FILE_NOT_FOUND", fmt.Sprintf("File not found: %s", f))
			os.Exit(1)
		}
	}

	// Create output directory
	os.MkdirAll("out", 0755)

	// Parse block files
	parsedBlocks, err := block.ProcessBlockFiles(blkPath, revPath, xorPath)
	if err != nil {
		writeError("PARSE_ERROR", err.Error())
		os.Exit(1)
	}

	// Run chain analysis
	blkFilename := filepath.Base(blkPath)
	result := analysis.AnalyzeBlocks(parsedBlocks, blkFilename)

	// Get stem name for output files
	stem := analysis.GetBlockStem(blkFilename)

	// Write JSON output
	jsonPath := filepath.Join("out", stem+".json")
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		writeError("JSON_ERROR", err.Error())
		os.Exit(1)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		writeError("WRITE_ERROR", err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Wrote %s (%d bytes)\n", jsonPath, len(jsonData))

	// Write Markdown report
	mdPath := filepath.Join("out", stem+".md")
	mdContent := report.GenerateMarkdownReport(result)
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		writeError("WRITE_ERROR", err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Wrote %s (%d bytes)\n", mdPath, len(mdContent))

	fmt.Fprintf(os.Stderr, "Chain analysis complete for %s\n", blkFilename)
}

func writeError(code, message string) {
	errResult := map[string]interface{}{
		"ok": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
	output, _ := json.Marshal(errResult)
	fmt.Println(string(output))
}
