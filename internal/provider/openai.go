package provider

import (
	"github.com/tmc/langchaingo/llms/openai"
)

// OpenAIProvider implements Provider for OpenAI
type OpenAIProvider struct {
	BaseProvider
}

// Initialize sets up the OpenAI provider
func (p *OpenAIProvider) Initialize(cfg Config) error {
	p.name = "openai"
	p.displayName = "OpenAI"
	p.config = cfg

	model := cfg.Model
	if model == "" {
		model = "gpt-4o"
	}

	opts := []openai.Option{
		openai.WithModel(model),
	}

	if cfg.APIKey != "" {
		opts = append(opts, openai.WithToken(cfg.APIKey))
	}

	if cfg.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.BaseURL))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return err
	}

	p.llm = llm
	return nil
}

// SupportsToolCalling returns true as OpenAI supports function calling
func (p *OpenAIProvider) SupportsToolCalling() bool {
	return true
}
