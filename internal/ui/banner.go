package ui

import (
	"fmt"
	"strings"

	gofigure "github.com/common-nighthawk/go-figure"
	"github.com/charmbracelet/lipgloss"
)

// PrintBanner prints the markdocs ASCII logo block to stdout.
// Call at the start of every command.
func PrintBanner(version string) {
	fmt.Println(Banner(version))
}

// Banner returns the rendered header block as a string.
func Banner(version string) string {
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted))

	// Generate ASCII art logo
	fig := gofigure.NewFigure("markdocs", "doom", true)
	lines := fig.Slicify()

	// Trim trailing whitespace from each line
	var trimmed []string
	for _, l := range lines {
		trimmed = append(trimmed, strings.TrimRight(l, " "))
	}
	logo := strings.Join(trimmed, "\n")

	return logoStyle.Render(logo) + "\n\n" +
		mutedStyle.Render("  compile docs into Claude skills") + "\n" +
		mutedStyle.Render("  " + version)
}
