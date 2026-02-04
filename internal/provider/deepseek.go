package provider

import (
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	deepseekBaseURL     = "https://api.deepseek.com/v1"
	deepseekDefaultModel = "deepseek-chat"
)

// DeepseekProvider implements Provider for DeepSeek (OpenAI-compatible)
type DeepseekProvider struct {
	BaseProvider
}

// Initialize sets up the DeepSeek provider
func (p *DeepseekProvider) Initialize(cfg Config) error {
	p.name = "deepseek"
	p.displayName = "DeepSeek"
	p.config = cfg

	model := cfg.Model
	if model == "" {
		model = deepseekDefaultModel
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = deepseekBaseURL
	}

	opts := []openai.Option{
		openai.WithModel(model),
		openai.WithBaseURL(baseURL),
	}

	if cfg.APIKey != "" {
		opts = append(opts, openai.WithToken(cfg.APIKey))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return err
	}

	p.llm = llm
	return nil
}

// SupportsToolCalling returns true as DeepSeek supports function calling
func (p *DeepseekProvider) SupportsToolCalling() bool {
	return true
}
