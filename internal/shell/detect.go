package shell

import (
	"os"
	"strings"

	"github.com/mitchellh/go-ps"
)

// Shell type constants
const (
	PowerShell = "powershell"
	Cmd        = "cmd"
	Bash       = "bash"
	Zsh        = "zsh"
	Fish       = "fish"
	Unknown    = "unknown"
)

// DetectParentShell detects the parent shell process
func DetectParentShell() string {
	pid := os.Getpid()
	proc, err := ps.FindProcess(pid)
	if err != nil || proc == nil {
		return Unknown
	}

	for i := 0; i < 10; i++ {
		proc, err = ps.FindProcess(proc.PPid())
		if err != nil || proc == nil {
			break
		}

		name := strings.ToLower(proc.Executable())
		switch {
		case strings.Contains(name, "powershell") || strings.Contains(name, "pwsh"):
			return PowerShell
		case name == "cmd.exe":
			return Cmd
		case strings.Contains(name, "bash"):
			return Bash
		case strings.Contains(name, "zsh"):
			return Zsh
		case strings.Contains(name, "fish"):
			return Fish
		}
	}

	return Unknown
}

// NormalizeShellName normalizes shell names to standard values
func NormalizeShellName(shell string) string {
	shell = strings.ToLower(strings.TrimSpace(shell))

	switch {
	case shell == "powershell" || shell == "pwsh" || shell == "ps":
		return PowerShell
	case shell == "cmd" || shell == "command":
		return Cmd
	case shell == "bash":
		return Bash
	case shell == "zsh":
		return Zsh
	case shell == "fish":
		return Fish
	default:
		return shell // Return as-is if not recognized
	}
}

// DetermineShell determines the shell to use based on overrides and detection
func DetermineShell(argOverride, envOverride string) string {
	// Priority: 1) Command line argument, 2) Environment variable, 3) Auto-detection
	if argOverride != "" {
		return NormalizeShellName(argOverride)
	}

	if envOverride != "" {
		return NormalizeShellName(envOverride)
	}

	return DetectParentShell()
}

// IsWindowsShell returns true if the shell is a Windows shell
func IsWindowsShell(shell string) bool {
	return shell == PowerShell || shell == Cmd
}

// IsUnixShell returns true if the shell is a Unix-like shell
func IsUnixShell(shell string) bool {
	return shell == Bash || shell == Zsh || shell == Fish
}

// GetShellDisplayName returns a human-readable shell name
func GetShellDisplayName(shell string) string {
	names := map[string]string{
		PowerShell: "PowerShell",
		Cmd:        "Command Prompt",
		Bash:       "Bash",
		Zsh:        "Zsh",
		Fish:       "Fish",
		Unknown:    "Unknown Shell",
	}

	if name, ok := names[shell]; ok {
		return name
	}
	return shell
}
