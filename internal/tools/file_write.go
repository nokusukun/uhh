package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileWriteTool implements a file writing tool
type FileWriteTool struct{}

// FileWriteInput represents the input for the file write tool
type FileWriteInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Append  bool   `json:"append,omitempty"`
}

// NewFileWriteTool creates a new file write tool
func NewFileWriteTool() *FileWriteTool {
	return &FileWriteTool{}
}

// Name returns the tool name
func (t *FileWriteTool) Name() string {
	return "file_write"
}

// Description returns the tool description
func (t *FileWriteTool) Description() string {
	return "Write content to a file. Input should be a JSON object with 'path' and 'content' fields. Optionally set 'append' to true to append instead of overwrite."
}

// Parameters returns the JSON schema for the tool parameters
func (t *FileWriteTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to write to the file",
			},
			"append": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, append to the file instead of overwriting",
			},
		},
		"required": []string{"path", "content"},
	}
}

// Execute writes content to a file
func (t *FileWriteTool) Execute(ctx context.Context, input Input) (Output, error) {
	var writeInput FileWriteInput

	if err := json.Unmarshal([]byte(input.Raw), &writeInput); err != nil {
		return NewErrorOutput(fmt.Errorf("invalid input: %w", err)), nil
	}

	if writeInput.Path == "" {
		return NewErrorOutputString("path cannot be empty"), nil
	}

	// Resolve path
	path := writeInput.Path
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

	// Create parent directories if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewErrorOutput(fmt.Errorf("failed to create directory: %w", err)), nil
	}

	// Check if file exists for backup
	var existed bool
	if _, err := os.Stat(path); err == nil {
		existed = true
	}

	// Write file
	var err error
	if writeInput.Append {
		f, openErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if openErr != nil {
			return NewErrorOutput(openErr), nil
		}
		_, err = f.WriteString(writeInput.Content)
		f.Close()
	} else {
		err = os.WriteFile(path, []byte(writeInput.Content), 0644)
	}

	if err != nil {
		return NewErrorOutput(err), nil
	}

	// Build result message
	action := "created"
	if existed {
		if writeInput.Append {
			action = "appended to"
		} else {
			action = "overwrote"
		}
	}

	return NewOutput(fmt.Sprintf("Successfully %s file: %s (%d bytes)", action, writeInput.Path, len(writeInput.Content))), nil
}

// RequiresConfirmation returns true as writing modifies the filesystem
func (t *FileWriteTool) RequiresConfirmation() bool {
	return true
}

// SafetyLevel returns moderate as this modifies files
func (t *FileWriteTool) SafetyLevel() SafetyLevel {
	return SafetyLevelModerate
}

// GetWriteDescription returns a human-readable description of the write operation
func GetWriteDescription(input FileWriteInput) string {
	action := "Write"
	if input.Append {
		action = "Append"
	}
	return fmt.Sprintf("%s %d bytes to %s", action, len(input.Content), input.Path)
}
