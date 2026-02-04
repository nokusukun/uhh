package provider

import (
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/llms/openai"
)

const (
	kimiBaseURL        = "https://api.moonshot.cn/v1"
	kimiCodingBaseURL  = "https://api.kimi.com/coding/v1"
	kimiDefaultModel   = "moonshot-v1-8k"
	kimiCodingModel    = "kimi-coding/k2p5"
)

// KimiProvider implements Provider for Kimi/Moonshot (OpenAI-compatible)
type KimiProvider struct {
	BaseProvider
}

// codingAgentTransport adds the required User-Agent header for KimiCoding API
type codingAgentTransport struct {
	base http.RoundTripper
}

func (t *codingAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "claude-code/1.0")
	return t.base.RoundTrip(req)
}

// Initialize sets up the Kimi provider
func (p *KimiProvider) Initialize(cfg Config) error {
	p.name = "kimi"
	p.displayName = "Kimi (Moonshot)"
	p.config = cfg

	model := cfg.Model
	if model == "" {
		model = kimiDefaultModel
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		// Use KimiCoding API if using a kimi-coding model or sk-kimi key
		if strings.HasPrefix(model, "kimi-coding") || strings.HasPrefix(cfg.APIKey, "sk-kimi-") {
			baseURL = kimiCodingBaseURL
		} else {
			baseURL = kimiBaseURL
		}
	}

	opts := []openai.Option{
		openai.WithModel(model),
		openai.WithBaseURL(baseURL),
	}

	if cfg.APIKey != "" {
		opts = append(opts, openai.WithToken(cfg.APIKey))
	}

	// Add custom HTTP client with User-Agent header for KimiCoding API
	if strings.Contains(baseURL, "api.kimi.com") || strings.HasPrefix(cfg.APIKey, "sk-kimi-") {
		httpClient := &http.Client{
			Transport: &codingAgentTransport{
				base: http.DefaultTransport,
			},
		}
		opts = append(opts, openai.WithHTTPClient(httpClient))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return err
	}

	p.llm = llm
	return nil
}

// SupportsToolCalling returns false for KimiCoding API (tool calling format not fully compatible)
func (p *KimiProvider) SupportsToolCalling() bool {
	// KimiCoding API has compatibility issues with standard tool calling format
	// Disable for now to prevent errors
	return false
}
