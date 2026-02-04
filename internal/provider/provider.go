package provider

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// Provider represents an LLM provider that can generate completions
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "gemini")
	Name() string

	// DisplayName returns human-readable name (e.g., "OpenAI", "Google Gemini")
	DisplayName() string

	// Initialize sets up the provider with the given config
	Initialize(cfg Config) error

	// LLM returns the underlying langchaingo LLM interface
	LLM() llms.Model

	// SupportsToolCalling returns whether this provider supports function calling
	SupportsToolCalling() bool

	// Call makes a simple text completion call
	Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error)

	// GenerateContent makes a content generation call with messages
	GenerateContent(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (*llms.ContentResponse, error)
}

// Config contains configuration for a provider
type Config struct {
	APIKey      string
	Model       string
	BaseURL     string
	Temperature float64
	MaxTokens   int
}

// BaseProvider provides common functionality for providers
type BaseProvider struct {
	name        string
	displayName string
	llm         llms.Model
	config      Config
}

// Name returns the provider identifier
func (p *BaseProvider) Name() string {
	return p.name
}

// DisplayName returns human-readable name
func (p *BaseProvider) DisplayName() string {
	return p.displayName
}

// LLM returns the underlying LLM model
func (p *BaseProvider) LLM() llms.Model {
	return p.llm
}

// Call makes a simple text completion call
func (p *BaseProvider) Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, p.llm, prompt, opts...)
}

// GenerateContent makes a content generation call with messages
func (p *BaseProvider) GenerateContent(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (*llms.ContentResponse, error) {
	return p.llm.GenerateContent(ctx, messages, opts...)
}
