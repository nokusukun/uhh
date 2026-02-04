package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"uhh/internal/agent"
	"uhh/internal/config"
	"uhh/internal/history"
	"uhh/internal/output"
	"uhh/internal/provider"
	"uhh/internal/shell"
	"uhh/internal/tools"
	"uhh/internal/tui"
	"uhh/internal/updater"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/llms"
)

// Version is set via ldflags at build time
var version = "dev"

var (
	// Flags
	providerFlag   string
	shellFlag      string
	modelFlag      string
	autoApproveFlag bool
	agentModeFlag  bool

	// Root command
	rootCmd = &cobra.Command{
		Use:   "uhh [prompt]",
		Short: "AI-powered terminal command assistant",
		Long: `UHH is an AI-powered CLI tool that helps you generate shell commands
from natural language descriptions. It supports multiple LLM providers
and can execute commands with tool calling.`,
		Args: cobra.ArbitraryArgs,
		Run:  runMain,
	}

	// Init command for onboarding
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize or reconfigure UHH",
		Long:  "Run the setup wizard to configure LLM providers and settings.",
		Run:   runInit,
	}

	// Config command
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Show current configuration",
		Run:   runConfigShow,
	}

	// Config models subcommand
	configModelsCmd = &cobra.Command{
		Use:   "models [provider]",
		Short: "List or select models for a provider",
		Long:  "List available models from the API and optionally select one. If no provider is specified, uses the default provider.",
		Run:   runConfigModels,
	}

	// Config set subcommand
	configSetCmd = &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  "Set a configuration value. Available keys: provider, model, auto-approve",
		Args:  cobra.ExactArgs(2),
		Run:   runConfigSet,
	}

	// Update command
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Check for and install updates",
		Run:   runUpdate,
	}

	// Version command
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("uhh version %s\n", version)
		},
	}
)

func init() {
	// Initialize colors
	output.InitColors()

	// Add flags to root command
	rootCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "LLM provider to use (openai, gemini, deepseek, kimi, glm)")
	rootCmd.Flags().StringVarP(&shellFlag, "shell", "s", "", "Override shell detection (powershell, cmd, bash, zsh, fish)")
	rootCmd.Flags().StringVarP(&modelFlag, "model", "m", "", "Model to use")
	rootCmd.Flags().BoolVarP(&autoApproveFlag, "auto-approve", "y", false, "Auto-approve tool executions")
	rootCmd.Flags().BoolVarP(&agentModeFlag, "agent", "a", false, "Run in agent mode with tool calling")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(versionCmd)

	// Add config subcommands
	configCmd.AddCommand(configModelsCmd)
	configCmd.AddCommand(configSetCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		output.PrintError(err.Error())
		os.Exit(1)
	}
}

func runMain(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	// Apply flag overrides
	if cfg.UI.NoColor {
		output.DisableColors()
	}

	// Check if first run
	if !config.Exists() {
		output.PrintInfo("First time setup detected. Running configuration wizard...")
		runInit(cmd, args)
		// Reload config after init
		cfg, err = config.Load()
		if err != nil {
			output.PrintError(fmt.Sprintf("Failed to load config: %v", err))
			os.Exit(1)
		}
	}

	// Get user prompt
	userPrompt := strings.Join(args, " ")
	if userPrompt == "" {
		output.PrintPrompt("What do you want? ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			userPrompt = scanner.Text()
		}
	}

	if userPrompt == "" {
		output.PrintWarn("No prompt provided. Exiting.")
		os.Exit(1)
	}

	// Determine provider
	providerName := providerFlag
	if providerName == "" {
		providerName = cfg.GetActiveProvider()
	}

	// Get provider settings
	providerSettings, ok := cfg.GetProviderSettings(providerName)
	if !ok {
		output.PrintError(fmt.Sprintf("Unknown provider: %s", providerName))
		os.Exit(1)
	}

	// Initialize provider
	p, err := provider.GetAndInitialize(providerName, provider.Config{
		APIKey:      providerSettings.APIKey,
		Model:       getModelOrDefault(providerSettings.Model, providerName),
		BaseURL:     providerSettings.BaseURL,
		Temperature: providerSettings.Temperature,
	})
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to initialize provider: %v", err))
		os.Exit(1)
	}

	// Determine shell
	shellName := shell.DetermineShell(shellFlag, cfg.Shell.Override)

	// Handle "actually" rewrite
	if strings.HasPrefix(strings.ToLower(userPrompt), "actually") {
		addendum := strings.TrimSpace(userPrompt[len("actually"):])
		lastPrompt, lastShell := history.LoadLastEntry()
		if lastPrompt != "" {
			if shellFlag == "" && cfg.Shell.Override == "" {
				shellName = lastShell
			}
			userPrompt = lastPrompt + ". " + addendum
			output.PrintInfo("Revising previous prompt with new info...")
		} else {
			output.PrintWarn("No history found for revision.")
		}
	}

	// Determine if we should use agent mode
	useAgent := agentModeFlag || (cfg.Agent.EnabledTools != nil && len(cfg.Agent.EnabledTools) > 0 && p.SupportsToolCalling())

	var completion string

	if useAgent && p.SupportsToolCalling() {
		// Agent mode with tool calling
		completion, err = runAgentMode(ctx, p, cfg, userPrompt, shellName)
	} else {
		// Simple mode
		completion, err = runSimpleMode(ctx, p, cfg, userPrompt, shellName)
	}

	if err != nil {
		output.PrintError(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}

	// Output result
	output.PrintCommand(completion)

	// Copy to clipboard
	if err := clipboard.WriteAll(completion); err == nil {
		output.PrintSuccess("Copied to clipboard!")
	}

	// Log history
	history.Log(shellName, userPrompt, completion)
}

func runSimpleMode(ctx context.Context, p provider.Provider, cfg *config.Config, userPrompt, shellName string) (string, error) {
	// Build prompt
	prompt := shell.BuildPrompt(userPrompt, shellName, cfg.Shell.AppendFileContext, cfg.Shell.MaxContextTokens)

	// Call LLM
	return p.Call(ctx, prompt, llms.WithTemperature(cfg.Providers[cfg.DefaultProvider].Temperature))
}

func runAgentMode(ctx context.Context, p provider.Provider, cfg *config.Config, userPrompt, shellName string) (string, error) {
	// Create tool registry
	toolRegistry := tools.DefaultRegistry()

	// Create agent
	agentConfig := agent.Config{
		AutoApprove:   autoApproveFlag || cfg.Agent.AutoApprove,
		MaxIterations: cfg.Agent.MaxIterations,
		Temperature:   cfg.Providers[cfg.DefaultProvider].Temperature,
	}

	a := agent.New(p, toolRegistry, agentConfig)

	// Set system prompt
	a.SetSystemPrompt(shell.BuildAgentSystemPrompt(shellName))

	// Set confirmation function if not auto-approve
	if !agentConfig.AutoApprove {
		a.SetConfirmFunc(tui.ConfirmToolExecution)
	}

	// Run agent
	result, err := a.Run(ctx, userPrompt)
	if err != nil {
		return "", err
	}

	// Print tool usage summary
	if len(result.ToolsUsed) > 0 {
		output.PrintDim(fmt.Sprintf("Used %d tools in %d iterations", len(result.ToolsUsed), result.Iterations))
	}

	return result.FinalAnswer, nil
}

func runInit(cmd *cobra.Command, args []string) {
	// Run onboarding wizard
	result, err := tui.RunOnboarding()
	if err != nil {
		output.PrintError(fmt.Sprintf("Setup failed: %v", err))
		os.Exit(1)
	}

	// Load or create config
	cfg := config.DefaultConfig()

	// Apply onboarding result
	tui.ApplyOnboardingResult(cfg, result)

	// Save config
	if err := cfg.Save(); err != nil {
		output.PrintError(fmt.Sprintf("Failed to save config: %v", err))
		os.Exit(1)
	}

	// Print success
	tui.PrintOnboardingSuccess(cfg)
}

func runConfigShow(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	fmt.Printf("Default Provider: %s\n", cfg.DefaultProvider)
	fmt.Printf("Auto-Approve: %v\n", cfg.Agent.AutoApprove)
	fmt.Printf("Max Iterations: %d\n", cfg.Agent.MaxIterations)
	fmt.Printf("Enabled Tools: %v\n", cfg.Agent.EnabledTools)
	fmt.Println()
	fmt.Println("Providers:")
	for name, settings := range cfg.Providers {
		status := "disabled"
		if settings.Enabled {
			status = "enabled"
		}
		hasKey := "no key"
		if settings.APIKey != "" {
			hasKey = "key set"
		}
		fmt.Printf("  %s: %s (%s, model: %s)\n", name, status, hasKey, settings.Model)
	}
}

func getModelOrDefault(model, providerName string) string {
	if model != "" {
		return model
	}
	if modelFlag != "" {
		return modelFlag
	}
	if defaultModel, ok := config.DefaultModels[providerName]; ok {
		return defaultModel
	}
	return "gpt-4o"
}

func runConfigModels(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	// Determine which provider to use
	providerName := cfg.DefaultProvider
	if len(args) > 0 {
		providerName = args[0]
	}

	// Check if provider exists
	settings, ok := cfg.Providers[providerName]
	if !ok {
		output.PrintError(fmt.Sprintf("Unknown provider: %s", providerName))
		os.Exit(1)
	}

	displayName := config.ProviderDisplayNames[providerName]
	output.PrintInfo(fmt.Sprintf("Fetching models for %s...", displayName))

	// List available models
	models, err := tui.ListAvailableModels(cfg, providerName)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to fetch models: %v", err))
		os.Exit(1)
	}

	// Print available models
	fmt.Printf("\nAvailable models for %s:\n", displayName)
	fmt.Printf("Current model: %s\n\n", settings.Model)
	for _, m := range models {
		marker := "  "
		if m.ID == settings.Model {
			marker = "* "
		}
		if m.Description != "" {
			fmt.Printf("%s%s - %s\n", marker, m.ID, m.Description)
		} else {
			fmt.Printf("%s%s\n", marker, m.ID)
		}
	}

	// Ask if user wants to change the model
	fmt.Print("\nDo you want to change the model? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != "yes" {
		return
	}

	// Run model selection
	selectedModel, err := tui.RunModelSelection(cfg, providerName)
	if err != nil {
		output.PrintError(fmt.Sprintf("Model selection failed: %v", err))
		os.Exit(1)
	}

	// Update config
	settings.Model = selectedModel
	cfg.Providers[providerName] = settings

	if err := cfg.Save(); err != nil {
		output.PrintError(fmt.Sprintf("Failed to save config: %v", err))
		os.Exit(1)
	}

	output.PrintSuccess(fmt.Sprintf("Model changed to: %s", selectedModel))
}

func runConfigSet(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	key := strings.ToLower(args[0])
	value := args[1]

	switch key {
	case "provider", "default_provider", "default-provider":
		// Validate provider exists
		if _, ok := cfg.Providers[value]; !ok {
			output.PrintError(fmt.Sprintf("Unknown provider: %s", value))
			os.Exit(1)
		}
		cfg.DefaultProvider = value
		output.PrintSuccess(fmt.Sprintf("Default provider set to: %s", value))

	case "model":
		// Set model for default provider
		providerName := cfg.DefaultProvider
		if settings, ok := cfg.Providers[providerName]; ok {
			settings.Model = value
			cfg.Providers[providerName] = settings
			output.PrintSuccess(fmt.Sprintf("Model for %s set to: %s", providerName, value))
		}

	case "auto-approve", "auto_approve", "autoapprove":
		val := strings.ToLower(value)
		cfg.Agent.AutoApprove = val == "true" || val == "1" || val == "yes"
		output.PrintSuccess(fmt.Sprintf("Auto-approve set to: %v", cfg.Agent.AutoApprove))

	default:
		output.PrintError(fmt.Sprintf("Unknown config key: %s", key))
		output.PrintInfo("Available keys: provider, model, auto-approve")
		os.Exit(1)
	}

	if err := cfg.Save(); err != nil {
		output.PrintError(fmt.Sprintf("Failed to save config: %v", err))
		os.Exit(1)
	}
}

func runUpdate(cmd *cobra.Command, args []string) {
	output.PrintInfo(fmt.Sprintf("Current version: %s", version))
	output.PrintInfo("Checking for updates...")

	info, err := updater.CheckForUpdate(version)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to check for updates: %v", err))
		os.Exit(1)
	}

	if !info.HasUpdate {
		output.PrintSuccess("You're running the latest version!")
		return
	}

	output.PrintInfo(fmt.Sprintf("New version available: %s", info.LatestVersion))

	// Ask for confirmation
	fmt.Print("Do you want to update? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != "yes" {
		output.PrintInfo("Update cancelled.")
		return
	}

	output.PrintInfo("Downloading update...")
	if err := updater.PerformUpdate(info); err != nil {
		output.PrintError(fmt.Sprintf("Failed to update: %v", err))
		os.Exit(1)
	}

	output.PrintSuccess(fmt.Sprintf("Successfully updated to %s!", info.LatestVersion))
	output.PrintInfo("Please restart uhh to use the new version.")
}
