package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	// DefaultProvider is the provider to use when none specified
	DefaultProvider string `json:"default_provider"`

	// Providers contains configuration for each provider
	Providers map[string]ProviderSettings `json:"providers"`

	// Agent settings
	Agent AgentSettings `json:"agent"`

	// Shell settings
	Shell ShellSettings `json:"shell"`

	// UI settings
	UI UISettings `json:"ui"`
}

// ProviderSettings contains settings for a single provider
type ProviderSettings struct {
	Enabled     bool    `json:"enabled"`
	APIKey      string  `json:"api_key,omitempty"`
	Model       string  `json:"model,omitempty"`
	BaseURL     string  `json:"base_url,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

// AgentSettings contains agent-specific configuration
type AgentSettings struct {
	AutoApprove   bool     `json:"auto_approve"`
	MaxIterations int      `json:"max_iterations"`
	EnabledTools  []string `json:"enabled_tools"`
}

// ShellSettings contains shell-related configuration
type ShellSettings struct {
	Override          string `json:"override,omitempty"`
	AppendFileContext bool   `json:"append_file_context"`
	MaxContextTokens  int    `json:"max_context_tokens"`
}

// UISettings contains UI preferences
type UISettings struct {
	NoColor     bool   `json:"no_color"`
	Theme       string `json:"theme"`
	ShowSpinner bool   `json:"show_spinner"`
}

// ConfigDir returns the path to the config directory
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".uhh"), nil
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads config from disk or returns defaults
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	return cfg, nil
}

// Save writes config to disk
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// Exists returns true if a config file exists
func Exists() bool {
	path, err := ConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// applyEnvOverrides applies environment variable overrides to config
func (c *Config) applyEnvOverrides() {
	// Provider override
	if provider := os.Getenv("UHH_PROVIDER"); provider != "" {
		c.DefaultProvider = provider
	}

	// API key overrides
	envKeys := map[string]string{
		"openai":   "OPENAI_API_KEY",
		"gemini":   "GOOGLE_API_KEY",
		"deepseek": "DEEPSEEK_API_KEY",
		"kimi":     "MOONSHOT_API_KEY",
		"glm":      "GLM_API_KEY",
	}

	for provider, envVar := range envKeys {
		if key := os.Getenv(envVar); key != "" {
			if settings, ok := c.Providers[provider]; ok {
				settings.APIKey = key
				c.Providers[provider] = settings
			}
		}
	}

	// Model override
	if model := os.Getenv("UHH_MODEL"); model != "" {
		if settings, ok := c.Providers[c.DefaultProvider]; ok {
			settings.Model = model
			c.Providers[c.DefaultProvider] = settings
		}
	}

	// Shell override
	if shell := os.Getenv("UHH_SHELL"); shell != "" {
		c.Shell.Override = shell
	}

	// No color override
	if os.Getenv("UHH_NO_COLOR") != "" || os.Getenv("NO_COLOR") != "" {
		c.UI.NoColor = true
	}

	// Auto approve override
	if autoApprove := os.Getenv("UHH_AUTO_APPROVE"); autoApprove != "" {
		c.Agent.AutoApprove = autoApprove == "1" || autoApprove == "true"
	}

	// File context override
	if appendContext := os.Getenv("UHH_APPEND_SMALL_CONTEXT"); appendContext != "" {
		if appendContext == "true" || appendContext == "1" {
			c.Shell.AppendFileContext = true
		} else if tokens, err := strconv.Atoi(appendContext); err == nil && tokens > 0 {
			c.Shell.AppendFileContext = true
			c.Shell.MaxContextTokens = tokens
		}
	}
}

// GetProviderSettings returns settings for the specified provider with env overrides
func (c *Config) GetProviderSettings(provider string) (ProviderSettings, bool) {
	settings, ok := c.Providers[provider]
	return settings, ok
}

// GetActiveProvider returns the currently active provider name
func (c *Config) GetActiveProvider() string {
	if provider := os.Getenv("UHH_PROVIDER"); provider != "" {
		return provider
	}
	return c.DefaultProvider
}
