package tui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// ConfirmToolExecution asks the user to confirm a tool execution
func ConfirmToolExecution(toolName, description, command string) (bool, error) {
	var confirmed bool

	// Build a description of what will be executed
	displayText := fmt.Sprintf("Tool: %s\n", toolName)
	if description != "" {
		displayText += fmt.Sprintf("Description: %s\n", description)
	}
	if command != "" {
		displayText += fmt.Sprintf("Command: %s", command)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Tool Execution Request").
				Description(displayText),
			huh.NewConfirm().
				Title("Allow this tool to execute?").
				Affirmative("Yes, execute").
				Negative("No, skip").
				Value(&confirmed),
		),
	).WithTheme(GetTheme())

	if err := form.Run(); err != nil {
		return false, err
	}

	return confirmed, nil
}

// ConfirmDangerousOperation asks for confirmation before a dangerous operation
func ConfirmDangerousOperation(operation, warning string) (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(FormatWarning("Warning: Dangerous Operation")).
				Description(warning),
			huh.NewConfirm().
				Title(fmt.Sprintf("Proceed with: %s?", operation)).
				Affirmative("Yes, I understand the risks").
				Negative("No, cancel").
				Value(&confirmed),
		),
	).WithTheme(GetTheme())

	if err := form.Run(); err != nil {
		return false, err
	}

	return confirmed, nil
}

// SimpleConfirm asks a simple yes/no question
func SimpleConfirm(question string) (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(question).
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed),
		),
	).WithTheme(GetTheme())

	if err := form.Run(); err != nil {
		return false, err
	}

	return confirmed, nil
}

// SelectOption presents options to the user and returns the selected one
func SelectOption(title string, options []string) (string, error) {
	var selected string

	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, opt)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(huhOptions...).
				Value(&selected),
		),
	).WithTheme(GetTheme())

	if err := form.Run(); err != nil {
		return "", err
	}

	return selected, nil
}
