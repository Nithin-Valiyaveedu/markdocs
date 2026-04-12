package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	lgtable "github.com/charmbracelet/lipgloss/table"
)

// Table wraps lipgloss/table for consistent markdocs styling.
type Table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a new table with the given column headers.
func NewTable(headers ...string) *Table {
	return &Table{headers: headers}
}

// Row adds a row to the table.
func (t *Table) Row(cols ...string) *Table {
	t.rows = append(t.rows, cols)
	return t
}

// Print renders and prints the table to stdout.
func (t *Table) Print() {
	fmt.Println(t.Render())
}

// Render returns the table as a string.
func (t *Table) Render() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Padding(0, 1)

	dimCellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Padding(0, 1)

	tbl := lgtable.New().
		Headers(t.headers...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBorder))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == lgtable.HeaderRow {
				return headerStyle
			}
			if row%2 == 0 {
				return dimCellStyle
			}
			return cellStyle
		})

	for _, r := range t.rows {
		tbl = tbl.Row(r...)
	}

	return tbl.String()
}
