package tui

import (
	"context"
	"fmt"
	"time"

	"uhh/internal/config"
	"uhh/internal/provider"

	"github.com/charmbracelet/huh"
)

// ProviderOption represents a provider option in the selection form
type ProviderOption struct {
	Name        string
	DisplayName string
	EnvVar      string
}

// Available providers for onboarding
var availableProviders = []ProviderOption{
	{Name: config.ProviderOpenAI, DisplayName: "OpenAI (GPT-4, GPT-4o)", EnvVar: "OPENAI_API_KEY"},
	{Name: config.ProviderGemini, DisplayName: "Google Gemini", EnvVar: "GOOGLE_API_KEY"},
	{Name: config.ProviderDeepseek, DisplayName: "DeepSeek", EnvVar: "DEEPSEEK_API_KEY"},
	{Name: config.ProviderKimi, DisplayName: "Kimi (Moonshot)", EnvVar: "MOONSHOT_API_KEY"},
	{Name: config.ProviderGLM, DisplayName: "GLM (Zhipu AI)", EnvVar: "GLM_API_KEY"},
}

// OnboardingResult contains the result of the onboarding wizard
type OnboardingResult struct {
	SelectedProviders []string
	APIKeys           map[string]string
	Models            map[string]string
	DefaultProvider   string
	AutoApprove       bool
}

// RunOnboarding runs the onboarding wizard and returns the configuration
func RunOnboarding() (*OnboardingResult, error) {
	result := &OnboardingResult{
		APIKeys: make(map[string]string),
		Models:  make(map[string]string),
	}

	// Welcome screen
	fmt.Println(FormatTitle("Welcome to UHH!"))
	fmt.Println(FormatSubtitle("Let's set up your AI providers.\n"))

	// Step 1: Provider selection
	var selectedProviders []string
	providerOptions := make([]huh.Option[string], len(availableProviders))
	for i, p := range availableProviders {
		providerOptions[i] = huh.NewOption(p.DisplayName, p.Name)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select the AI providers you want to configure").
				Description("You can configure multiple providers and switch between them.").
				Options(providerOptions...).
				Value(&selectedProviders),
		).Title("Step 1: Choose Providers"),
	).WithTheme(GetTheme())

	if err := form.Run(); err != nil {
		return nil, err
	}

	if len(selectedProviders) == 0 {
		return nil, fmt.Errorf("no providers selected")
	}

	result.SelectedProviders = selectedProviders

	// Step 2: API Keys and Model Selection for each selected provider
	for _, providerName := range selectedProviders {
		var providerOpt ProviderOption
		for _, p := range availableProviders {
			if p.Name == providerName {
				providerOpt = p
				break
			}
		}

		var apiKey string
		keyForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Enter your %s API key", providerOpt.DisplayName)).
					Description(fmt.Sprintf("Or set the %s environment variable later.", providerOpt.EnvVar)).
					Placeholder("sk-...").
					EchoMode(huh.EchoModePassword).
					Value(&apiKey),
			).Title(fmt.Sprintf("Step 2: %s Configuration", providerOpt.DisplayName)),
		).WithTheme(GetTheme())

		if err := keyForm.Run(); err != nil {
			return nil, err
		}

		if apiKey != "" {
			result.APIKeys[providerName] = apiKey

			// Try to fetch available models
			selectedModel, err := selectModelForProvider(providerName, providerOpt.DisplayName, apiKey)
			if err == nil && selectedModel != "" {
				result.Models[providerName] = selectedModel
			}
		}
	}

	// Step 3: Default provider selection
	if len(selectedProviders) > 1 {
		defaultOptions := make([]huh.Option[string], len(selectedProviders))
		for i, name := range selectedProviders {
			displayName := name
			for _, p := range availableProviders {
				if p.Name == name {
					displayName = p.DisplayName
					break
				}
			}
			defaultOptions[i] = huh.NewOption(displayName, name)
		}

		defaultForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Which provider should be the default?").
					Options(defaultOptions...).
					Value(&result.DefaultProvider),
			).Title("Step 3: Default Provider"),
		).WithTheme(GetTheme())

		if err := defaultForm.Run(); err != nil {
			return nil, err
		}
	} else {
		result.DefaultProvider = selectedProviders[0]
	}

	// Step 4: Auto-approve setting
	settingsForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Auto-approve tool executions?").
				Description("If enabled, tools will execute without asking for confirmation. Not recommended for beginners.").
				Affirmative("Yes, auto-approve").
				Negative("No, always ask").
				Value(&result.AutoApprove),
		).Title("Step 4: Settings"),
	).WithTheme(GetTheme())

	if err := settingsForm.Run(); err != nil {
		return nil, err
	}

	return result, nil
}

// ApplyOnboardingResult applies the onboarding result to a config
func ApplyOnboardingResult(cfg *config.Config, result *OnboardingResult) {
	cfg.DefaultProvider = result.DefaultProvider
	cfg.Agent.AutoApprove = result.AutoApprove

	// Enable selected providers and set API keys and models
	for _, name := range result.SelectedProviders {
		if settings, ok := cfg.Providers[name]; ok {
			settings.Enabled = true
			if key, hasKey := result.APIKeys[name]; hasKey && key != "" {
				settings.APIKey = key
			}
			if model, hasModel := result.Models[name]; hasModel && model != "" {
				settings.Model = model
			}
			cfg.Providers[name] = settings
		}
	}

	// Disable non-selected providers
	for name := range cfg.Providers {
		found := false
		for _, selected := range result.SelectedProviders {
			if name == selected {
				found = true
				break
			}
		}
		if !found {
			settings := cfg.Providers[name]
			settings.Enabled = false
			cfg.Providers[name] = settings
		}
	}
}

// selectModelForProvider fetches available models and lets the user select one
func selectModelForProvider(providerName, displayName, apiKey string) (string, error) {
	fmt.Printf("%s Fetching available models...\n", FormatDim("→"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to fetch models from API
	models, err := provider.ListModels(ctx, providerName, apiKey, "")

	// Fall back to defaults if API fetch fails
	if err != nil || len(models) == 0 {
		fmt.Printf("%s Using default model list\n", FormatDim("→"))
		models = provider.GetDefaultModelsForProvider(providerName)
	}

	if len(models) == 0 {
		return "", fmt.Errorf("no models available")
	}

	// Build options for selection
	options := make([]huh.Option[string], 0, len(models))
	for _, m := range models {
		label := m.ID
		if m.Name != "" && m.Name != m.ID {
			label = fmt.Sprintf("%s (%s)", m.Name, m.ID)
		}
		if m.Description != "" && len(m.Description) < 50 {
			label = fmt.Sprintf("%s - %s", label, m.Description)
		}
		options = append(options, huh.NewOption(label, m.ID))
	}

	// Get default model for this provider
	defaultModel := config.DefaultModels[providerName]

	var selectedModel string
	// Try to pre-select the default model
	for _, m := range models {
		if m.ID == defaultModel {
			selectedModel = defaultModel
			break
		}
	}
	if selectedModel == "" && len(models) > 0 {
		selectedModel = models[0].ID
	}

	modelForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Select model for %s", displayName)).
				Description(fmt.Sprintf("Found %d available models", len(models))).
				Options(options...).
				Value(&selectedModel),
		),
	).WithTheme(GetTheme())

	if err := modelForm.Run(); err != nil {
		return "", err
	}

	return selectedModel, nil
}

// PrintOnboardingSuccess prints a success message after onboarding
func PrintOnboardingSuccess(cfg *config.Config) {
	fmt.Println()
	fmt.Println(FormatSuccess("Configuration saved successfully!"))
	fmt.Println()
	defaultSettings := cfg.Providers[cfg.DefaultProvider]
	fmt.Printf("Default provider: %s\n", FormatInfo(config.ProviderDisplayNames[cfg.DefaultProvider]))
	fmt.Printf("Model: %s\n", FormatInfo(defaultSettings.Model))
	fmt.Printf("Auto-approve: %s\n", FormatInfo(fmt.Sprintf("%v", cfg.Agent.AutoApprove)))
	fmt.Println()
	fmt.Println(FormatDim("Run 'uhh <your prompt>' to get started!"))
	fmt.Println(FormatDim("Run 'uhh init' to reconfigure at any time."))
	fmt.Println(FormatDim("Run 'uhh config models' to change the model."))
}

// RunModelSelection runs a model selection wizard for a specific provider
func RunModelSelection(cfg *config.Config, providerName string) (string, error) {
	settings, ok := cfg.Providers[providerName]
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", providerName)
	}

	displayName := config.ProviderDisplayNames[providerName]
	apiKey := settings.APIKey

	if apiKey == "" {
		return "", fmt.Errorf("no API key configured for %s", displayName)
	}

	return selectModelForProvider(providerName, displayName, apiKey)
}

// ListAvailableModels fetches and displays available models for a provider
func ListAvailableModels(cfg *config.Config, providerName string) ([]provider.ModelInfo, error) {
	settings, ok := cfg.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	apiKey := settings.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("no API key configured for %s", providerName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	models, err := provider.ListModels(ctx, providerName, apiKey, settings.BaseURL)
	if err != nil {
		// Fall back to defaults
		models = provider.GetDefaultModelsForProvider(providerName)
		if len(models) == 0 {
			return nil, err
		}
	}

	return models, nil
}
