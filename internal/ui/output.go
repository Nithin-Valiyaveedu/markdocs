package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	symbolCheck   = "✓"
	symbolCross   = "✗"
	symbolWarning = "○"
	symbolArrow   = "↻"
	symbolDot     = "·"
)

// Step prints a numbered step indicator (without completion status).
// Use before starting an async operation, then call StepDone when it finishes.
func Step(n int, msg string) {
	badge := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true).
		Render(fmt.Sprintf("  %d", n))
	fmt.Printf("%s  %s\n", badge, StyleMuted.Render(msg))
}

// StepDone prints a completed step with a green checkmark.
func StepDone(n int, msg string) {
	_ = n // reserved for future aligned output
	badge := StyleSuccess.Render("  " + symbolCheck)
	fmt.Printf("%s  %s\n", badge, msg)
}

// Success prints a green checkmark line.
func Success(msg string) {
	fmt.Printf("%s  %s\n", StyleSuccess.Render("  "+symbolCheck), msg)
}

// Warning prints an amber circle line.
func Warning(msg string) {
	fmt.Printf("%s  %s\n", StyleWarning.Render("  "+symbolWarning), msg)
}

// Error prints a red cross line.
func Error(msg string) {
	fmt.Printf("%s  %s\n", StyleError.Render("  "+symbolCross), msg)
}

// Info prints a muted indented line with no status symbol.
func Info(msg string) {
	fmt.Printf("     %s\n", StyleMuted.Render(msg))
}

// Recompiling prints an amber recompile indicator.
func Recompiling(msg string) {
	fmt.Printf("%s  %s\n", StyleWarning.Render("  "+symbolArrow), msg)
}

// Section prints a labeled horizontal divider.
func Section(title string) {
	width := 44
	line := strings.Repeat("─", width-len(title)-2)
	styled := StyleMuted.Render("── " + title + " " + line)
	fmt.Println(styled)
}

// Blank prints an empty line.
func Blank() {
	fmt.Println()
}
