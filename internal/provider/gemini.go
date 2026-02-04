package provider

import (
	"context"

	"github.com/tmc/langchaingo/llms/googleai"
)

// GeminiProvider implements Provider for Google Gemini
type GeminiProvider struct {
	BaseProvider
}

// Initialize sets up the Gemini provider
func (p *GeminiProvider) Initialize(cfg Config) error {
	p.name = "gemini"
	p.displayName = "Google Gemini"
	p.config = cfg

	model := cfg.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	opts := []googleai.Option{
		googleai.WithDefaultModel(model),
	}

	if cfg.APIKey != "" {
		opts = append(opts, googleai.WithAPIKey(cfg.APIKey))
	}

	if cfg.Temperature > 0 {
		opts = append(opts, googleai.WithDefaultTemperature(cfg.Temperature))
	}

	if cfg.MaxTokens > 0 {
		opts = append(opts, googleai.WithDefaultMaxTokens(cfg.MaxTokens))
	}

	llm, err := googleai.New(context.Background(), opts...)
	if err != nil {
		return err
	}

	p.llm = llm
	return nil
}

// SupportsToolCalling returns true as Gemini supports function calling
func (p *GeminiProvider) SupportsToolCalling() bool {
	return true
}
