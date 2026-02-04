package tui

import (
	"fmt"

	"uhh/internal/config"

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
	DefaultProvider   string
	AutoApprove       bool
}

// RunOnboarding runs the onboarding wizard and returns the configuration
func RunOnboarding() (*OnboardingResult, error) {
	result := &OnboardingResult{
		APIKeys: make(map[string]string),
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

	// Step 2: API Keys for each selected provider
	for _, providerName := range selectedProviders {
		var provider ProviderOption
		for _, p := range availableProviders {
			if p.Name == providerName {
				provider = p
				break
			}
		}

		var apiKey string
		keyForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Enter your %s API key", provider.DisplayName)).
					Description(fmt.Sprintf("Or set the %s environment variable later.", provider.EnvVar)).
					Placeholder("sk-...").
					EchoMode(huh.EchoModePassword).
					Value(&apiKey),
			).Title(fmt.Sprintf("Step 2: %s API Key", provider.DisplayName)),
		).WithTheme(GetTheme())

		if err := keyForm.Run(); err != nil {
			return nil, err
		}

		if apiKey != "" {
			result.APIKeys[providerName] = apiKey
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

	// Enable selected providers and set API keys
	for _, name := range result.SelectedProviders {
		if settings, ok := cfg.Providers[name]; ok {
			settings.Enabled = true
			if key, hasKey := result.APIKeys[name]; hasKey && key != "" {
				settings.APIKey = key
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

// PrintOnboardingSuccess prints a success message after onboarding
func PrintOnboardingSuccess(cfg *config.Config) {
	fmt.Println()
	fmt.Println(FormatSuccess("Configuration saved successfully!"))
	fmt.Println()
	fmt.Printf("Default provider: %s\n", FormatInfo(config.ProviderDisplayNames[cfg.DefaultProvider]))
	fmt.Printf("Auto-approve: %s\n", FormatInfo(fmt.Sprintf("%v", cfg.Agent.AutoApprove)))
	fmt.Println()
	fmt.Println(FormatDim("Run 'uhh <your prompt>' to get started!"))
	fmt.Println(FormatDim("Run 'uhh init' to reconfigure at any time."))
}
