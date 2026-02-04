package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	maxOutputLen   = 10000
)

// BashTool implements a shell command execution tool
type BashTool struct {
	Timeout time.Duration
}

// BashInput represents the input for the bash tool
type BashInput struct {
	Command string `json:"command"`
}

// NewBashTool creates a new bash tool with default settings
func NewBashTool() *BashTool {
	return &BashTool{
		Timeout: defaultTimeout,
	}
}

// Name returns the tool name
func (t *BashTool) Name() string {
	return "bash"
}

// Description returns the tool description
func (t *BashTool) Description() string {
	return "Execute shell commands. Use this to run terminal commands and scripts. Input should be a JSON object with a 'command' field."
}

// Parameters returns the JSON schema for the tool parameters
func (t *BashTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
		},
		"required": []string{"command"},
	}
}

// Execute runs the bash command
func (t *BashTool) Execute(ctx context.Context, input Input) (Output, error) {
	var bashInput BashInput

	// Try to parse as JSON first
	if err := json.Unmarshal([]byte(input.Raw), &bashInput); err != nil {
		// If not valid JSON, treat the raw input as the command
		bashInput.Command = input.Raw
	}

	if bashInput.Command == "" {
		return NewErrorOutputString("command cannot be empty"), nil
	}

	// Check for dangerous commands
	if warning := checkDangerousCommand(bashInput.Command); warning != "" {
		// Still execute but include warning
		// In agent mode, the confirmation prompt handles this
	}

	// Set up context with timeout
	execCtx, cancel := context.WithTimeout(ctx, t.Timeout)
	defer cancel()

	// Determine shell and execute
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/C", bashInput.Command)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", bashInput.Command)
	}

	// Set working directory if specified
	if input.WorkingDir != "" {
		cmd.Dir = input.WorkingDir
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	err := cmd.Run()

	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded {
		return NewErrorOutputString(fmt.Sprintf("command timed out after %v", t.Timeout)), nil
	}

	// Build result
	result := stdout.String()
	if stderr.Len() > 0 {
		if result != "" {
			result += "\n"
		}
		result += "[stderr]\n" + stderr.String()
	}

	// Truncate if too long
	if len(result) > maxOutputLen {
		result = result[:maxOutputLen] + "\n... (output truncated)"
	}

	if err != nil {
		if result == "" {
			return NewErrorOutput(err), nil
		}
		// Include the output even if there was an error
		return Output{
			Success: false,
			Result:  result,
			Error:   err.Error(),
		}, nil
	}

	if result == "" {
		result = "(command completed with no output)"
	}

	return NewOutput(result), nil
}

// RequiresConfirmation returns true as bash commands can be dangerous
func (t *BashTool) RequiresConfirmation() bool {
	return true
}

// SafetyLevel returns dangerous as this tool executes arbitrary commands
func (t *BashTool) SafetyLevel() SafetyLevel {
	return SafetyLevelDangerous
}

// Dangerous command patterns
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\s+(-rf?|--recursive)\s*/`),
	regexp.MustCompile(`(?i)\brm\s+-rf?\s+\*`),
	regexp.MustCompile(`(?i)\bsudo\s+rm`),
	regexp.MustCompile(`(?i)\bmkfs\b`),
	regexp.MustCompile(`(?i)\bdd\s+.*of=/dev/`),
	regexp.MustCompile(`(?i)>\s*/dev/sd[a-z]`),
	regexp.MustCompile(`(?i)\bchmod\s+777`),
	regexp.MustCompile(`(?i)\bcurl\s+.*\|\s*(ba)?sh`),
	regexp.MustCompile(`(?i)\bwget\s+.*\|\s*(ba)?sh`),
	regexp.MustCompile(`(?i):(){ :|:& };:`), // Fork bomb
	regexp.MustCompile(`(?i)\bformat\s+[a-z]:`), // Windows format
}

// checkDangerousCommand checks if a command matches dangerous patterns
func checkDangerousCommand(command string) string {
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(command) {
			return fmt.Sprintf("Warning: This command matches a dangerous pattern. Please review carefully.")
		}
	}
	return ""
}

// IsDangerousCommand returns true if the command matches dangerous patterns
func IsDangerousCommand(command string) bool {
	return checkDangerousCommand(command) != ""
}

// GetDangerousCommandWarning returns a warning if the command is dangerous
func GetDangerousCommandWarning(command string) string {
	// Check against patterns
	warning := checkDangerousCommand(command)
	if warning != "" {
		return warning
	}

	// Additional checks for common dangerous operations
	lowerCmd := strings.ToLower(command)

	if strings.Contains(lowerCmd, "rm ") && strings.Contains(lowerCmd, " -r") {
		return "Warning: Recursive delete operation detected."
	}

	if strings.Contains(lowerCmd, "sudo ") {
		return "Warning: Command requires elevated privileges."
	}

	if strings.Contains(lowerCmd, "> /") || strings.Contains(lowerCmd, ">> /") {
		return "Warning: Writing to system paths detected."
	}

	return ""
}
