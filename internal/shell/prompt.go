package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const promptTemplate = `<instruction>
You are a autocorrect system for a terminal, your environment is %s. When presented an input you fix and/or change it into a compatible %s command that can be executed.
</instruction>
%s<user_input>
%s
</user_input>
<output>
Only output a command that can be immediately executed.
DO NOT wrap in code blocks or anything else.
</output>`

// BuildPrompt creates the LLM prompt with shell context
func BuildPrompt(query, shell string, appendContext bool, maxTokens int) string {
	contextSection := ""
	if appendContext {
		contextSection = buildFileContext(query, maxTokens)
	}

	return fmt.Sprintf(promptTemplate, shell, shell, contextSection, query)
}

// BuildAgentSystemPrompt creates the system prompt for agent mode
func BuildAgentSystemPrompt(shell string) string {
	return fmt.Sprintf(`You are an AI assistant that helps users with terminal commands and tasks.
Your environment is %s. You have access to tools that can execute commands and read/write files.

When the user asks for help:
1. Analyze their request
2. Use the available tools to accomplish the task
3. Explain what you're doing and the results

Always prefer using tools over just providing text responses when actions are needed.
Be careful with destructive operations - confirm with the user if uncertain.`, shell)
}

// buildFileContext builds the file context section from referenced files
func buildFileContext(query string, maxTokens int) string {
	files := ExtractFileReferences(query)
	if len(files) == 0 {
		return ""
	}

	var contextFiles []string
	var contextFileNames []string

	for _, file := range files {
		if small, err := IsSmallFile(file, maxTokens); err == nil && small {
			content, err := os.ReadFile(file)
			if err == nil {
				contextFiles = append(contextFiles, string(content))
				contextFileNames = append(contextFileNames, file)
			}
		}
	}

	if len(contextFiles) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<file_contexts>\n")
	for i, content := range contextFiles {
		sb.WriteString(fmt.Sprintf("<file name='%s'>\n", filepath.Base(contextFileNames[i])))
		sb.WriteString(content)
		sb.WriteString("\n</file>\n")
	}
	sb.WriteString("</file_contexts>\n")

	return sb.String()
}

// ExtractFileReferences finds potential file paths in the user prompt
func ExtractFileReferences(text string) []string {
	var files []string
	seen := make(map[string]bool)

	patterns := []string{
		// Files with extensions
		`\b[\w\-\.\/\\]+\.[a-zA-Z0-9]+\b`,
		// Quoted file paths
		`["']([^"']+\.[a-zA-Z0-9]+)["']`,
		// Common config files
		`\b(package\.json|go\.mod|go\.sum|Dockerfile|Makefile|README\.md|\.gitignore)\b`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		for _, match := range matches {
			match = strings.Trim(match, `"'`)
			if !seen[match] {
				files = append(files, match)
				seen[match] = true
			}
		}
	}

	return files
}

// IsSmallFile checks if a file exists and is small enough to include
func IsSmallFile(filePath string, maxTokens int) (bool, error) {
	// Convert relative paths to absolute
	if !filepath.IsAbs(filePath) {
		abs, err := filepath.Abs(filePath)
		if err == nil {
			filePath = abs
		}
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	// Rough approximation: 1 token ~ 4 characters
	maxBytes := int64(maxTokens * 4)
	return info.Size() <= maxBytes, nil
}
