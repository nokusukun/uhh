package config

// Provider type constants
const (
	ProviderOpenAI   = "openai"
	ProviderGemini   = "gemini"
	ProviderDeepseek = "deepseek"
	ProviderKimi     = "kimi"
	ProviderGLM      = "glm"
)

// Default model names for each provider
var DefaultModels = map[string]string{
	ProviderOpenAI:   "gpt-4o",
	ProviderGemini:   "gemini-2.0-flash",
	ProviderDeepseek: "deepseek-chat",
	ProviderKimi:     "kimi-coding/k2p5",
	ProviderGLM:      "glm-4",
}

// Default base URLs for OpenAI-compatible providers
var DefaultBaseURLs = map[string]string{
	ProviderDeepseek: "https://api.deepseek.com/v1",
	ProviderKimi:     "https://api.moonshot.cn/v1",
	ProviderGLM:      "https://open.bigmodel.cn/api/paas/v4",
}

// Provider display names
var ProviderDisplayNames = map[string]string{
	ProviderOpenAI:   "OpenAI",
	ProviderGemini:   "Google Gemini",
	ProviderDeepseek: "DeepSeek",
	ProviderKimi:     "Kimi (Moonshot)",
	ProviderGLM:      "GLM (Zhipu AI)",
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		DefaultProvider: ProviderOpenAI,
		Providers: map[string]ProviderSettings{
			ProviderOpenAI: {
				Enabled:     true,
				Model:       DefaultModels[ProviderOpenAI],
				Temperature: 0.7,
			},
			ProviderGemini: {
				Enabled:     false,
				Model:       DefaultModels[ProviderGemini],
				Temperature: 0.7,
			},
			ProviderDeepseek: {
				Enabled:     false,
				Model:       DefaultModels[ProviderDeepseek],
				BaseURL:     DefaultBaseURLs[ProviderDeepseek],
				Temperature: 0.7,
			},
			ProviderKimi: {
				Enabled:     false,
				Model:       DefaultModels[ProviderKimi],
				BaseURL:     DefaultBaseURLs[ProviderKimi],
				Temperature: 0.7,
			},
			ProviderGLM: {
				Enabled:     false,
				Model:       DefaultModels[ProviderGLM],
				BaseURL:     DefaultBaseURLs[ProviderGLM],
				Temperature: 0.7,
			},
		},
		Agent: AgentSettings{
			AutoApprove:   false,
			MaxIterations: 10,
			EnabledTools:  []string{"bash", "file_read", "file_write"},
		},
		Shell: ShellSettings{
			Override:          "",
			AppendFileContext: false,
			MaxContextTokens:  1000,
		},
		UI: UISettings{
			NoColor:     false,
			Theme:       "charm",
			ShowSpinner: true,
		},
	}
}

// AllProviders returns a list of all available provider names
func AllProviders() []string {
	return []string{
		ProviderOpenAI,
		ProviderGemini,
		ProviderDeepseek,
		ProviderKimi,
		ProviderGLM,
	}
}
