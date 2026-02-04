package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"uhh/internal/provider"
	"uhh/internal/tools"

	"github.com/tmc/langchaingo/llms"
)

// ConfirmFunc is called to confirm tool execution
type ConfirmFunc func(toolName, description, command string) (bool, error)

// Config contains agent configuration
type Config struct {
	AutoApprove   bool
	MaxIterations int
	Temperature   float64
}

// DefaultConfig returns default agent configuration
func DefaultConfig() Config {
	return Config{
		AutoApprove:   false,
		MaxIterations: 10,
		Temperature:   0.7,
	}
}

// Agent represents an AI agent with tool-calling capabilities
type Agent struct {
	provider  provider.Provider
	tools     *tools.Registry
	config    Config
	context   *Context
	confirmFn ConfirmFunc
}

// ToolExecution records a tool execution
type ToolExecution struct {
	ToolName  string
	Input     string
	Output    string
	Approved  bool
	Skipped   bool
	Duration  time.Duration
	Error     error
}

// Result contains the result of an agent run
type Result struct {
	FinalAnswer    string
	ToolsUsed      []ToolExecution
	Iterations     int
	Success        bool
	Error          error
}

// New creates a new agent
func New(p provider.Provider, t *tools.Registry, cfg Config) *Agent {
	return &Agent{
		provider: p,
		tools:    t,
		config:   cfg,
		context:  NewContext(""),
	}
}

// SetConfirmFunc sets the confirmation function for dangerous operations
func (a *Agent) SetConfirmFunc(fn ConfirmFunc) {
	a.confirmFn = fn
}

// SetSystemPrompt sets the system prompt for the agent
func (a *Agent) SetSystemPrompt(prompt string) {
	a.context.SystemPrompt = prompt
}

// Run executes the agent loop with the given user prompt
func (a *Agent) Run(ctx context.Context, userPrompt string) (*Result, error) {
	result := &Result{
		ToolsUsed:  make([]ToolExecution, 0),
		Iterations: 0,
	}

	// Add user message to context
	a.context.AddUserMessage(userPrompt)

	// Get available tools
	availableTools := a.tools.ToLangchainTools()

	for i := 0; i < a.config.MaxIterations; i++ {
		result.Iterations = i + 1

		// Build messages
		messages := a.context.ToLangchainMessages()

		// Call LLM with tools
		opts := []llms.CallOption{
			llms.WithTemperature(a.config.Temperature),
		}

		if len(availableTools) > 0 {
			opts = append(opts, llms.WithTools(availableTools))
		}

		response, err := a.provider.GenerateContent(ctx, messages, opts...)
		if err != nil {
			result.Error = err
			return result, err
		}

		if len(response.Choices) == 0 {
			result.Error = fmt.Errorf("no response from LLM")
			return result, result.Error
		}

		choice := response.Choices[0]

		// Check if there are tool calls
		if len(choice.ToolCalls) > 0 {
			// Add assistant message with tool calls to context FIRST
			a.context.AddAssistantMessageWithToolCalls(choice.Content, choice.ToolCalls)

			// Process tool calls
			for _, toolCall := range choice.ToolCalls {
				execution := a.executeToolCall(ctx, toolCall)
				result.ToolsUsed = append(result.ToolsUsed, execution)

				// Add tool result to context
				if execution.Error != nil {
					a.context.AddToolResult(toolCall.ID, fmt.Sprintf("Error: %v", execution.Error))
				} else if execution.Skipped {
					a.context.AddToolResult(toolCall.ID, "Tool execution was skipped by user.")
				} else {
					a.context.AddToolResult(toolCall.ID, execution.Output)
				}
			}
			continue // Continue the loop to get next response
		}

		// No tool calls - this is the final answer
		if choice.Content != "" {
			result.FinalAnswer = choice.Content
			result.Success = true
			a.context.AddAssistantMessage(choice.Content)
			return result, nil
		}

		// Check stop reason
		if choice.StopReason == "end_turn" || choice.StopReason == "stop" {
			result.FinalAnswer = choice.Content
			result.Success = true
			return result, nil
		}
	}

	result.Error = fmt.Errorf("max iterations (%d) reached", a.config.MaxIterations)
	return result, result.Error
}

// executeToolCall executes a single tool call
func (a *Agent) executeToolCall(ctx context.Context, toolCall llms.ToolCall) ToolExecution {
	execution := ToolExecution{
		ToolName: toolCall.FunctionCall.Name,
		Input:    toolCall.FunctionCall.Arguments,
	}

	start := time.Now()

	// Get the tool
	tool, err := a.tools.Get(toolCall.FunctionCall.Name)
	if err != nil {
		execution.Error = err
		execution.Duration = time.Since(start)
		return execution
	}

	// Check if confirmation is needed
	needsConfirmation := tool.RequiresConfirmation() && !a.config.AutoApprove

	if needsConfirmation && a.confirmFn != nil {
		// Build description for confirmation
		description := fmt.Sprintf("Execute %s tool", tool.Name())

		approved, err := a.confirmFn(tool.Name(), description, toolCall.FunctionCall.Arguments)
		if err != nil {
			execution.Error = err
			execution.Duration = time.Since(start)
			return execution
		}

		if !approved {
			execution.Skipped = true
			execution.Duration = time.Since(start)
			return execution
		}
		execution.Approved = true
	} else {
		execution.Approved = true
	}

	// Parse input
	var parsed map[string]interface{}
	json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &parsed)

	input := tools.Input{
		Raw:    toolCall.FunctionCall.Arguments,
		Parsed: parsed,
	}

	// Execute tool
	output, err := tool.Execute(ctx, input)
	execution.Duration = time.Since(start)

	if err != nil {
		execution.Error = err
		return execution
	}

	if output.Success {
		execution.Output = output.Result
	} else {
		execution.Output = output.Error
		if output.Result != "" {
			execution.Output = output.Result + "\n" + output.Error
		}
	}

	return execution
}

// SimpleCall makes a simple LLM call without tool support
func (a *Agent) SimpleCall(ctx context.Context, prompt string) (string, error) {
	return a.provider.Call(ctx, prompt, llms.WithTemperature(a.config.Temperature))
}

// Reset clears the conversation context
func (a *Agent) Reset() {
	a.context.Clear()
}
