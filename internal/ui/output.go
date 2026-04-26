package ui

import (
	"github.com/fatih/color"
)

// PrintSuccess prints a success message in green.
func PrintSuccess(message string) {
	color.New(color.FgGreen).Printf("\033[1;32m✔ %s\033[0m\n", message)
}

// PrintError prints an error message in red.
func PrintError(message string) {
	color.New(color.FgRed).Printf("\033[1;31m✘ %s\033[0m\n", message)
}

// PrintInfo prints an informational message in blue.
func PrintInfo(message string) {
	color.New(color.FgBlue).Printf("\033[1;34mℹ %s\033[0m\n", message)
}

// PrintWarning prints a warning message in yellow.
func PrintWarning(message string) {
	color.New(color.FgYellow).Printf("\033[1;33m⚠ %s\033[0m\n", message)
}

// PrintLog prints a runtime log message in cyan.
func PrintLog(message string) {
	color.New(color.FgCyan).Printf("\033[1;36m• %s\033[0m\n", message)
}
