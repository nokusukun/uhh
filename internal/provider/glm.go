package provider

import (
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	glmBaseURL      = "https://open.bigmodel.cn/api/paas/v4"
	glmDefaultModel = "glm-4"
)

// GLMProvider implements Provider for GLM/Zhipu AI (OpenAI-compatible)
type GLMProvider struct {
	BaseProvider
}

// Initialize sets up the GLM provider
func (p *GLMProvider) Initialize(cfg Config) error {
	p.name = "glm"
	p.displayName = "GLM (Zhipu AI)"
	p.config = cfg

	model := cfg.Model
	if model == "" {
		model = glmDefaultModel
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = glmBaseURL
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

// SupportsToolCalling returns true as GLM supports function calling
func (p *GLMProvider) SupportsToolCalling() bool {
	return true
}
