package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

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
	// Dim text for secondary info
	dimColor = color.New(color.Faint)
	// Tool name in magenta
	toolColor = color.New(color.FgMagenta)
)

// DisableColors disables all color output
func DisableColors() {
	color.NoColor = true
}

// EnableColors enables color output
func EnableColors() {
	color.NoColor = false
}

// InitColors initializes color settings based on environment
func InitColors() {
	if os.Getenv("UHH_NO_COLOR") != "" || os.Getenv("NO_COLOR") != "" {
		DisableColors()
	}
}

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

// PrintPrompt prints prompt text in blue (no newline)
func PrintPrompt(msg string) {
	promptColor.Print(msg)
}

// PrintDim prints dimmed/secondary text
func PrintDim(msg string) {
	dimColor.Println(msg)
}

// PrintTool prints tool name in magenta
func PrintTool(name string) {
	toolColor.Printf("[%s] ", name)
}

// PrintToolOutput prints tool execution output
func PrintToolOutput(toolName, output string) {
	PrintTool(toolName)
	fmt.Println(output)
}

// PrintToolError prints tool error
func PrintToolError(toolName string, err error) {
	PrintTool(toolName)
	errorColor.Printf("Error: %v\n", err)
}

// PrintThinking prints a "thinking" indicator
func PrintThinking(msg string) {
	dimColor.Printf("... %s\n", msg)
}

// PrintConfirmation prints a confirmation prompt
func PrintConfirmation(msg string) {
	warnColor.Print(msg)
}

// Sprintf formats with color support
func Sprintf(c *color.Color, format string, a ...interface{}) string {
	return c.Sprintf(format, a...)
}

// CommandString returns a formatted command string
func CommandString(cmd string) string {
	return cmdColor.Sprint(cmd)
}

// ErrorString returns a formatted error string
func ErrorString(msg string) string {
	return errorColor.Sprint(msg)
}

// InfoString returns a formatted info string
func InfoString(msg string) string {
	return infoColor.Sprint(msg)
}
