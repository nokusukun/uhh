package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Colors for the TUI theme
var (
	Primary   = lipgloss.Color("#7C3AED") // Purple
	Secondary = lipgloss.Color("#06B6D4") // Cyan
	Success   = lipgloss.Color("#10B981") // Green
	Warning   = lipgloss.Color("#F59E0B") // Yellow
	Error     = lipgloss.Color("#EF4444") // Red
	Muted     = lipgloss.Color("#6B7280") // Gray
	Text      = lipgloss.Color("#F9FAFB") // White
	TextDim   = lipgloss.Color("#9CA3AF") // Light gray
)

// Styles for various UI elements
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextDim).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	DimStyle = lipgloss.NewStyle().
			Foreground(Muted)

	CodeStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)
)

// GetTheme returns the huh theme for forms
func GetTheme() *huh.Theme {
	return huh.ThemeCharm()
}

// FormatTitle formats a title string
func FormatTitle(title string) string {
	return TitleStyle.Render(title)
}

// FormatSubtitle formats a subtitle string
func FormatSubtitle(subtitle string) string {
	return SubtitleStyle.Render(subtitle)
}

// FormatSuccess formats a success message
func FormatSuccess(msg string) string {
	return SuccessStyle.Render(msg)
}

// FormatError formats an error message
func FormatError(msg string) string {
	return ErrorStyle.Render(msg)
}

// FormatWarning formats a warning message
func FormatWarning(msg string) string {
	return WarningStyle.Render(msg)
}

// FormatInfo formats an info message
func FormatInfo(msg string) string {
	return InfoStyle.Render(msg)
}

// FormatCode formats code/command text
func FormatCode(code string) string {
	return CodeStyle.Render(code)
}

// FormatDim formats dimmed/secondary text
func FormatDim(text string) string {
	return DimStyle.Render(text)
}

// FormatBox wraps text in a box
func FormatBox(content string) string {
	return BoxStyle.Render(content)
}
