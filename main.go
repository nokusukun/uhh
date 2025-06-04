package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/mitchellh/go-ps"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

func init() {
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

func main() {
	llm, err := openai.New(
		openai.WithModel("gpt-4o"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	userPrompt := GetUserPrompt()
	shell := DetectParentShell()

	// Handle "actually" rewrite
	if strings.HasPrefix(strings.ToLower(userPrompt), "actually") {
		addendum := strings.TrimSpace(userPrompt[len("actually"):])
		lastPrompt, lastShell := LoadLastPrompt()
		if lastPrompt != "" {
			shell = lastShell // prefer last shell if found
			userPrompt = lastPrompt + ". " + addendum
			fmt.Println("(Revising previous prompt with new info.)")
		} else {
			fmt.Println("No history found for revision.")
		}
	}

	fmt.Println("Detected shell:", shell)
	prompt := Prompt(userPrompt, shell)
	completion, err := llm.Call(ctx, prompt, llms.WithTemperature(1))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
	fmt.Println("Copied to clipboard!")
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
