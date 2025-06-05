package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/mitchellh/go-ps"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func init() {
	// Disable colors if requested via environment variable
	if os.Getenv("UHH_NO_COLOR") != "" || os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}

	// Set up OpenAI API key
	token := os.Getenv("OPENAI_API_KEY")
	if token != "" {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println("Warning: Failed to get user home directory:", err)
		return
	}
	file, err := os.ReadFile(path.Join(home, ".openai.token.txt"))
	if err != nil {
		log.Println("Warning: Failed to read token file:", err)
		return
	}
	fileToken := strings.TrimSpace(string(file))
	if fileToken == "" {
		log.Printf("Warning: No OpenAI API key found in environment variable or file.")
	}
	err = os.Setenv("OPENAI_API_KEY", fileToken)
	if err != nil {
		log.Println("Warning: Failed to set OpenAI API key from file:", err)
	}
}

func DetectParentShell() string {
	pid := os.Getpid()
	proc, err := ps.FindProcess(pid)
	if err != nil || proc == nil {
		return "unknown"
	}
	for i := 0; i < 10; i++ {
		proc, err = ps.FindProcess(proc.PPid())
		if err != nil || proc == nil {
			break
		}
		name := strings.ToLower(proc.Executable())
		switch {
		case strings.Contains(name, "powershell") || strings.Contains(name, "pwsh"):
			return "powershell"
		case name == "cmd.exe":
			return "cmd"
		case strings.Contains(name, "bash"):
			return "bash"
		case strings.Contains(name, "zsh"):
			return "zsh"
		case strings.Contains(name, "fish"):
			return "fish"
		}
	}
	return "unknown"
}

func Prompt(query string, shell string) string {
	p := `
<instruction>
You are a autocorrect system for a terminal, your environment is %shell%. When presented an input you fix and/or change it into a compatible %shell% command that can be executed.
</instruction>
<user_input>
%query%
</user_input>
<output>
Only output a command that can be immediately executed.
DO NOT wrap in code blocks or anything else.
</output>
`
	p = strings.ReplaceAll(p, "%shell%", shell)
	p = strings.ReplaceAll(p, "%query%", query)

	// Append small file context if enabled
	p = AppendSmallFileContext(p, query)

	return p
}

func GetHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./.uhh.history.txt"
	}
	return path.Join(home, ".uhh.history.txt")
}

func LogHistory(entry string) {
	histPath := GetHistoryPath()
	f, err := os.OpenFile(histPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		// Use regular log for internal warnings to avoid disrupting output flow
		log.Printf("Warning: Failed to write history: %v", err)
		return
	}
	defer f.Close()
	fmt.Fprintln(f, entry)
}

func LoadLastPrompt() (string, string) {
	histPath := GetHistoryPath()
	file, err := os.Open(histPath)
	if err != nil {
		return "", ""
	}
	defer file.Close()
	var lastPrompt, lastShell string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Prompt: ") {
			lastPrompt = strings.TrimPrefix(line, "Prompt: ")
		}
		if strings.HasPrefix(line, "Shell: ") {
			lastShell = strings.TrimPrefix(line, "Shell: ")
		}
	}
	return lastPrompt, lastShell
}

func GetUserPrompt() string {
	prompts := strings.Join(os.Args[1:], " ")
	if prompts == "" {
		fmt.Println("Please enter your prompt:")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			prompts = scanner.Text()
		}
	}
	if prompts == "" {
		fmt.Println("No prompt provided. Exiting.")
		os.Exit(1)
	}
	return prompts
}

// ParseShellOverride extracts shell override from arguments and returns cleaned prompt and shell
func ParseShellOverride(args []string) ([]string, string) {
	var cleanedArgs []string
	var shellOverride string

	for i, arg := range args {
		if strings.HasPrefix(arg, "!shell=") {
			// Handle !shell=cmd format
			shellOverride = strings.TrimPrefix(arg, "!shell=")
		} else if arg == "--shell" && i+1 < len(args) {
			// Handle --shell cmd format
			shellOverride = args[i+1]
			// Skip the next argument as it's the shell value
			i++
		} else {
			cleanedArgs = append(cleanedArgs, arg)
		}
	}

	return cleanedArgs, shellOverride
}

// GetUserPromptAndShell parses command line arguments for both prompt and shell override
func GetUserPromptAndShell() (string, string) {
	cleanedArgs, shellOverride := ParseShellOverride(os.Args[1:])

	prompts := strings.Join(cleanedArgs, " ")
	if prompts == "" {
		PrintPrompt("What do you want? ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			prompts = scanner.Text()
		}
	}
	if prompts == "" {
		PrintWarn("No prompt provided. Exiting.")
		os.Exit(1)
	}

	return prompts, shellOverride
}

// DetermineShell determines the shell to use based on overrides and detection
func DetermineShell(argShellOverride string) string {
	// Priority: 1) Command line argument, 2) Environment variable, 3) Auto-detection

	// Check command line argument first
	if argShellOverride != "" {
		return normalizeShellName(argShellOverride)
	}

	// Check environment variable
	if envShell := os.Getenv("UHH_SHELL"); envShell != "" {
		return normalizeShellName(envShell)
	}

	// Fall back to auto-detection
	return DetectParentShell()
}

// normalizeShellName normalizes shell names to standard values
func normalizeShellName(shell string) string {
	shell = strings.ToLower(strings.TrimSpace(shell))

	switch {
	case shell == "powershell" || shell == "pwsh" || shell == "ps":
		return "powershell"
	case shell == "cmd" || shell == "command":
		return "cmd"
	case shell == "bash":
		return "bash"
	case shell == "zsh":
		return "zsh"
	case shell == "fish":
		return "fish"
	default:
		return shell // Return as-is if not recognized
	}
}

// ExtractFileReferences finds potential file paths in the user prompt
func ExtractFileReferences(text string) []string {
	var files []string

	// Common file patterns
	patterns := []string{
		// Files with extensions
		`\b[\w\-\.\/\\]+\.[a-zA-Z0-9]+\b`,
		// Quoted file paths
		`["']([^"']+\.[a-zA-Z0-9]+)["']`,
		// Common config files without need for extensions
		`\b(package\.json|go\.mod|go\.sum|Dockerfile|Makefile|README\.md|\.gitignore)\b`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		for _, match := range matches {
			// Clean up quotes if present
			match = strings.Trim(match, `"'`)
			files = append(files, match)
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

	// Rough approximation: 1 token ≈ 4 characters
	maxBytes := int64(maxTokens * 4)
	return info.Size() <= maxBytes, nil
}

// AppendSmallFileContext adds content of small referenced files to the prompt
func AppendSmallFileContext(prompt, userPrompt string) string {
	appendContext := os.Getenv("UHH_APPEND_SMALL_CONTEXT")
	if appendContext == "" || strings.ToLower(appendContext) == "false" || appendContext == "0" {
		return prompt
	}

	maxTokens := 1000 // Default to 1000 tokens
	if appendContext != "true" && appendContext != "1" {
		// Try to parse as number
		if tokens := parseTokenLimit(appendContext); tokens > 0 {
			maxTokens = tokens
		}
	}

	files := ExtractFileReferences(userPrompt)
	var contextFiles []string
	var contextFileNames []string

	for _, file := range files {
		if small, err := IsSmallFile(file, maxTokens); err == nil && small {
			content, err := os.ReadFile(file)
			if err == nil {
				contextFiles = append(contextFiles, fmt.Sprintf("File: %s\n%s", file, string(content)))
				contextFileNames = append(contextFileNames, file)
			}
		}
	}

	if len(contextFiles) > 0 {
		contextSection := "<file_contexts>"
		for i, fileContent := range contextFiles {
			contextSection += "\n<file name='" + filepath.Base(contextFileNames[i]) + "'>\n"
			contextSection += fileContent + "\n"
			contextSection += "</file>\n"
		}
		contextSection += "</file_contexts>\n"
		return strings.Replace(prompt, "<user_input>", contextSection+"\n<user_input>", 1)
	}

	return prompt
}

// parseTokenLimit attempts to parse a string as a token limit
func parseTokenLimit(s string) int {
	// Simple parsing - just look for numbers
	re := regexp.MustCompile(`\d+`)
	if match := re.FindString(s); match != "" {
		var num int
		if n, err := fmt.Sscanf(match, "%d", &num); n == 1 && err == nil {
			return num
		}
	}
	return 0
}

func GetModel() string {
	model := "gpt-4o"
	if modelEnv := os.Getenv("UHH_MODEL"); modelEnv != "" {
		model = modelEnv
	}
	return model
}

// Color utility functions
var (
	// Command output in bright green
	cmdColor = color.New(color.FgHiGreen, color.Bold)
	// Success messages in green
	successColor = color.New(color.FgGreen)
	// Info messages in cyan
	infoColor = color.New(color.FgCyan)
	// Warning messages in yellow
	warnColor = color.New(color.FgYellow)
	// Error messages in red
	errorColor = color.New(color.FgRed, color.Bold)
	// Prompt text in blue
	promptColor = color.New(color.FgBlue)
)

// PrintCommand prints the generated command in bright green
func PrintCommand(cmd string) {
	cmdColor.Println(cmd)
}

// PrintSuccess prints success messages in green
func PrintSuccess(msg string) {
	successColor.Println(msg)
}

// PrintInfo prints informational messages in cyan
func PrintInfo(msg string) {
	infoColor.Println(msg)
}

// PrintWarn prints warning messages in yellow
func PrintWarn(msg string) {
	warnColor.Println(msg)
}

// PrintError prints error messages in red
func PrintError(msg string) {
	errorColor.Println(msg)
}

// PrintPrompt prints prompt text in blue
func PrintPrompt(msg string) {
	promptColor.Print(msg)
}

func main() {
	llm, err := openai.New(
		openai.WithModel(GetModel()),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	userPrompt, shellOverride := GetUserPromptAndShell()
	shell := DetermineShell(shellOverride)

	// Handle "actually" rewrite
	if strings.HasPrefix(strings.ToLower(userPrompt), "actually") {
		addendum := strings.TrimSpace(userPrompt[len("actually"):])
		lastPrompt, lastShell := LoadLastPrompt()
		if lastPrompt != "" {
			// Only use last shell if no override is specified
			if shellOverride == "" && os.Getenv("UHH_SHELL") == "" {
				shell = lastShell
			}
			userPrompt = lastPrompt + ". " + addendum
			PrintInfo("→ Revising previous prompt with new info...")
		} else {
			PrintWarn("No history found for revision.")
		}
	}

	// fmt.Println("Using shell:", shell)
	prompt := Prompt(userPrompt, shell)
	completion, err := llm.Call(ctx, prompt, llms.WithTemperature(1))
	if err != nil {
		PrintError(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
	PrintCommand(completion)
	PrintSuccess("✓ Copied to clipboard!")
	_ = clipboard.WriteAll(completion)

	// Log history
	histEntry := fmt.Sprintf(
		"Time: %s\nShell: %s\nPrompt: %s\nOutput: %s\n---",
		time.Now().Format(time.RFC3339),
		shell,
		userPrompt,
		completion,
	)
	LogHistory(histEntry)
}
