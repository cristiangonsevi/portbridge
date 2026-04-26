package ui

import (
	"fmt"

	"github.com/fatih/color"
)

// PrintSuccess prints a success message in green.
func PrintSuccess(message string) {
	color.New(color.FgGreen).Printf("✔ %s\n", message)
}

// PrintError prints an error message in red.
func PrintError(message string) {
	color.New(color.FgRed).Printf("✘ %s\n", message)
}

// PrintInfo prints an informational message in blue.
func PrintInfo(message string) {
	color.New(color.FgBlue).Printf("ℹ %s\n", message)
}

// PrintWarning prints a warning message in yellow.
func PrintWarning(message string) {
	color.New(color.FgYellow).Printf("⚠ %s\n", message)
}

// PrintLog prints a runtime log message in cyan.
func PrintLog(message string) {
	color.New(color.FgCyan).Printf("• %s\n", message)
}

// Printf is a generic formatted print helper using the default color.
func Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
