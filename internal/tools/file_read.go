package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxFileSize  = 100 * 1024 // 100KB max file size
	maxLineCount = 1000       // Max lines to read
)

// FileReadTool implements a file reading tool
type FileReadTool struct{}

// FileReadInput represents the input for the file read tool
type FileReadInput struct {
	Path string `json:"path"`
}

// NewFileReadTool creates a new file read tool
func NewFileReadTool() *FileReadTool {
	return &FileReadTool{}
}

// Name returns the tool name
func (t *FileReadTool) Name() string {
	return "file_read"
}

// Description returns the tool description
func (t *FileReadTool) Description() string {
	return "Read the contents of a file. Input should be a JSON object with a 'path' field containing the file path."
}

// Parameters returns the JSON schema for the tool parameters
func (t *FileReadTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to read",
			},
		},
		"required": []string{"path"},
	}
}

// Execute reads the file contents
func (t *FileReadTool) Execute(ctx context.Context, input Input) (Output, error) {
	var readInput FileReadInput

	if err := json.Unmarshal([]byte(input.Raw), &readInput); err != nil {
		// Try treating raw input as path
		readInput.Path = strings.TrimSpace(input.Raw)
	}

	if readInput.Path == "" {
		return NewErrorOutputString("path cannot be empty"), nil
	}

	// Resolve path
	path := readInput.Path
	if !filepath.IsAbs(path) {
		if input.WorkingDir != "" {
			path = filepath.Join(input.WorkingDir, path)
		} else {
			absPath, err := filepath.Abs(path)
			if err == nil {
				path = absPath
			}
		}
	}

	// Security check - prevent path traversal
	if strings.Contains(path, "..") {
		cleanPath := filepath.Clean(path)
		if strings.HasPrefix(cleanPath, "..") {
			return NewErrorOutputString("path traversal not allowed"), nil
		}
	}

	// Check file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewErrorOutputString(fmt.Sprintf("file not found: %s", readInput.Path)), nil
		}
		return NewErrorOutput(err), nil
	}

	// Check if it's a directory
	if info.IsDir() {
		return NewErrorOutputString(fmt.Sprintf("%s is a directory, not a file", readInput.Path)), nil
	}

	// Check file size
	if info.Size() > maxFileSize {
		return NewErrorOutputString(fmt.Sprintf("file too large (%d bytes, max %d bytes)", info.Size(), maxFileSize)), nil
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return NewErrorOutput(err), nil
	}

	// Truncate by line count if needed
	lines := strings.Split(string(content), "\n")
	if len(lines) > maxLineCount {
		content = []byte(strings.Join(lines[:maxLineCount], "\n") + "\n... (truncated)")
	}

	return NewOutput(string(content)), nil
}

// RequiresConfirmation returns false as reading is safe
func (t *FileReadTool) RequiresConfirmation() bool {
	return false
}

// SafetyLevel returns safe as this is a read-only operation
func (t *FileReadTool) SafetyLevel() SafetyLevel {
	return SafetyLevelSafe
}
