package history

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry represents a single history entry
type Entry struct {
	Time   time.Time
	Shell  string
	Prompt string
	Output string
}

// GetHistoryPath returns the path to the history file
func GetHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./.uhh.history.txt"
	}
	return filepath.Join(home, ".uhh.history.txt")
}

// LogEntry logs a history entry to the history file
func LogEntry(entry Entry) {
	histPath := GetHistoryPath()
	f, err := os.OpenFile(histPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Warning: Failed to write history: %v", err)
		return
	}
	defer f.Close()

	histEntry := fmt.Sprintf(
		"Time: %s\nShell: %s\nPrompt: %s\nOutput: %s\n---\n",
		entry.Time.Format(time.RFC3339),
		entry.Shell,
		entry.Prompt,
		entry.Output,
	)
	fmt.Fprint(f, histEntry)
}

// Log logs a simple history entry with the current time
func Log(shell, prompt, output string) {
	LogEntry(Entry{
		Time:   time.Now(),
		Shell:  shell,
		Prompt: prompt,
		Output: output,
	})
}

// LoadLastEntry loads the last prompt and shell from history
func LoadLastEntry() (prompt string, shell string) {
	histPath := GetHistoryPath()
	file, err := os.Open(histPath)
	if err != nil {
		return "", ""
	}
	defer file.Close()

	var lastPrompt, lastShell string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Prompt: ") {
			lastPrompt = strings.TrimPrefix(line, "Prompt: ")
		}
		if strings.HasPrefix(line, "Shell: ") {
			lastShell = strings.TrimPrefix(line, "Shell: ")
		}
	}

	return lastPrompt, lastShell
}

// LoadRecentEntries loads the N most recent history entries
func LoadRecentEntries(n int) []Entry {
	histPath := GetHistoryPath()
	file, err := os.Open(histPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var entries []Entry
	var current Entry
	var inEntry bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "Time: ") {
			if inEntry {
				entries = append(entries, current)
			}
			current = Entry{}
			inEntry = true
			timeStr := strings.TrimPrefix(line, "Time: ")
			if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
				current.Time = t
			}
		} else if strings.HasPrefix(line, "Shell: ") {
			current.Shell = strings.TrimPrefix(line, "Shell: ")
		} else if strings.HasPrefix(line, "Prompt: ") {
			current.Prompt = strings.TrimPrefix(line, "Prompt: ")
		} else if strings.HasPrefix(line, "Output: ") {
			current.Output = strings.TrimPrefix(line, "Output: ")
		} else if line == "---" {
			if inEntry {
				entries = append(entries, current)
				inEntry = false
			}
		}
	}

	// Return last n entries
	if len(entries) > n {
		return entries[len(entries)-n:]
	}
	return entries
}

// Clear clears the history file
func Clear() error {
	histPath := GetHistoryPath()
	return os.WriteFile(histPath, []byte{}, 0600)
}
