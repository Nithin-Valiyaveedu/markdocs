package ui

import "github.com/charmbracelet/lipgloss"

// Color palette — single source of truth.
const (
	ColorPrimary = "#7B61FF" // purple — brand
	ColorSuccess = "#04B575" // green
	ColorWarning = "#FFB627" // amber
	ColorError   = "#FF5F5F" // red
	ColorMuted   = "#6C6C6C" // gray
	ColorText    = "#DDDDDD" // off-white
	ColorBorder  = "#3D3D3D" // dark border
)

var (
	StylePrimary = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary)).Bold(true)
	StyleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess)).Bold(true)
	StyleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning))
	StyleError   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Bold(true)
	StyleMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted))
	StyleBold    = lipgloss.NewStyle().Bold(true)
	StyleText    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorText))
)
