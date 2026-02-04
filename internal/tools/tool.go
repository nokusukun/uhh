package tools

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// Tool represents an executable tool that the agent can use
type Tool interface {
	// Name returns the tool identifier
	Name() string

	// Description returns a description for the LLM to understand usage
	Description() string

	// Parameters returns JSON schema for the tool's parameters
	Parameters() map[string]interface{}

	// Execute runs the tool with given input
	Execute(ctx context.Context, input Input) (Output, error)

	// RequiresConfirmation returns whether user confirmation is needed
	RequiresConfirmation() bool

	// SafetyLevel returns the safety level of this tool
	SafetyLevel() SafetyLevel
}

// Input represents input to a tool
type Input struct {
	Raw        string                 // Raw input string (JSON)
	Parsed     map[string]interface{} // Parsed JSON input
	WorkingDir string                 // Current working directory
}

// Output represents output from a tool
type Output struct {
	Success bool   // Whether execution succeeded
	Result  string // The output/result
	Error   string // Error message if failed
}

// SafetyLevel defines tool execution safety levels
type SafetyLevel int

const (
	// SafetyLevelSafe - Read-only operations, no side effects
	SafetyLevelSafe SafetyLevel = iota
	// SafetyLevelModerate - May write files, needs confirmation by default
	SafetyLevelModerate
	// SafetyLevelDangerous - Executes arbitrary commands, always confirm
	SafetyLevelDangerous
)

// String returns a string representation of the safety level
func (s SafetyLevel) String() string {
	switch s {
	case SafetyLevelSafe:
		return "safe"
	case SafetyLevelModerate:
		return "moderate"
	case SafetyLevelDangerous:
		return "dangerous"
	default:
		return "unknown"
	}
}

// ToLangchainTool converts a Tool to langchaingo Tool format
func ToLangchainTool(t Tool) llms.Tool {
	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Parameters(),
		},
	}
}

// ToLangchainTools converts multiple tools to langchaingo format
func ToLangchainTools(tools []Tool) []llms.Tool {
	result := make([]llms.Tool, len(tools))
	for i, t := range tools {
		result[i] = ToLangchainTool(t)
	}
	return result
}

// NewOutput creates a successful output
func NewOutput(result string) Output {
	return Output{
		Success: true,
		Result:  result,
	}
}

// NewErrorOutput creates an error output
func NewErrorOutput(err error) Output {
	return Output{
		Success: false,
		Error:   err.Error(),
	}
}

// NewErrorOutputString creates an error output from a string
func NewErrorOutputString(errMsg string) Output {
	return Output{
		Success: false,
		Error:   errMsg,
	}
}
