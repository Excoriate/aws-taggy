package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Column defines a table column with its properties
type Column struct {
	Title    string
	Key      string
	Width    int
	Flexible bool
	Align    string // "left", "center", or "right"
}

// TableOptions contains configuration options for the table
type TableOptions struct {
	Title           string
	Columns         []Column
	Style           *TableStyle
	MaxHeight       int
	FlexibleColumns bool
	AutoWidth       bool
}

// TableStyle defines the styling for the TUI table
type TableStyle struct {
	BaseStyle   lipgloss.Style
	HeaderStyle lipgloss.Style
	RowStyle    lipgloss.Style
	SelectedRow lipgloss.Style
}

// DefaultTableStyle provides a clean, modern default styling
func DefaultTableStyle() TableStyle {
	return TableStyle{
		BaseStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")),
		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("240")).
			Bold(true),
		RowStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		SelectedRow: lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")),
	}
}

// calculateColumnWidths determines the optimal width for each column based on content
func calculateColumnWidths(columns []Column, data [][]string) []int {
	widths := make([]int, len(columns))

	// Initialize with header lengths
	for i, col := range columns {
		widths[i] = len(col.Title)
	}

	// Check content lengths
	for _, row := range data {
		for i, cell := range row {
			if i < len(widths) {
				cellWidth := len(cell)
				if cellWidth > widths[i] {
					widths[i] = cellWidth
				}
			}
		}
	}

	// Add padding
	for i := range widths {
		widths[i] += 2 // Add minimal padding
	}

	return widths
}

// padString ensures a string fits the given width with proper padding
func padString(s string, width int, align string) string {
	sLen := len(s)
	if sLen >= width {
		if align == "right" {
			return s[sLen-width:]
		}
		return s[:width]
	}

	spaces := width - sLen
	switch align {
	case "right":
		return strings.Repeat(" ", spaces) + s
	case "center":
		left := spaces / 2
		right := spaces - left
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	default: // left align
		return s + strings.Repeat(" ", spaces)
	}
}

// RenderTable creates a generic table for rendering data
func RenderTable(opts TableOptions, data [][]string) error {
	if len(opts.Columns) == 0 {
		return fmt.Errorf("no columns defined")
	}

	// Ensure all data rows have the correct number of columns
	normalizedData := make([][]string, len(data))
	for i, row := range data {
		normalizedRow := make([]string, len(opts.Columns))
		for j := range opts.Columns {
			if j < len(row) {
				normalizedRow[j] = strings.TrimSpace(row[j]) // Trim spaces to ensure clean alignment
			} else {
				normalizedRow[j] = "" // Fill missing columns with empty strings
			}
		}
		normalizedData[i] = normalizedRow
	}

	var columns []table.Column
	var columnWidths []int

	// Calculate column widths
	if opts.AutoWidth {
		columnWidths = calculateColumnWidths(opts.Columns, normalizedData)
	} else {
		columnWidths = make([]int, len(opts.Columns))
		for i, col := range opts.Columns {
			if col.Width > 0 {
				columnWidths[i] = col.Width
			} else {
				columnWidths[i] = 20 // Default width
			}
		}
	}

	// Create columns with proper alignment
	for i, col := range opts.Columns {
		width := columnWidths[i]
		if col.Width > 0 {
			width = col.Width // Use specified width for fixed-width columns
		}

		// Determine alignment for the column
		align := col.Align
		if align == "" {
			// Default alignments based on content type
			switch {
			case strings.Contains(strings.ToLower(col.Title), "count"),
				strings.Contains(strings.ToLower(col.Title), "size"),
				strings.Contains(strings.ToLower(col.Title), "number"):
				align = "right"
			case strings.Contains(strings.ToLower(col.Title), "status"),
				strings.Contains(strings.ToLower(col.Title), "tags"),
				strings.Contains(strings.ToLower(col.Title), "type"):
				align = "center"
			default:
				align = "left"
			}
		}

		columns = append(columns, table.Column{
			Title: padString(strings.TrimSpace(col.Title), width, align),
			Width: width,
		})
	}

	// Convert string slices to table.Row with proper padding and alignment
	rows := make([]table.Row, len(normalizedData))
	for i, rowData := range normalizedData {
		paddedRow := make([]string, len(rowData))
		for j, cell := range rowData {
			align := opts.Columns[j].Align
			if align == "" {
				// Use same alignment logic as headers
				switch {
				case strings.Contains(strings.ToLower(opts.Columns[j].Title), "count"),
					strings.Contains(strings.ToLower(opts.Columns[j].Title), "size"),
					strings.Contains(strings.ToLower(opts.Columns[j].Title), "number"):
					align = "right"
				case strings.Contains(strings.ToLower(opts.Columns[j].Title), "status"),
					strings.Contains(strings.ToLower(opts.Columns[j].Title), "tags"),
					strings.Contains(strings.ToLower(opts.Columns[j].Title), "type"):
					align = "center"
				default:
					align = "left"
				}
			}

			width := columnWidths[j]
			if opts.Columns[j].Width > 0 {
				width = opts.Columns[j].Width // Use specified width for fixed-width columns
			}

			paddedRow[j] = padString(strings.TrimSpace(cell), width, align)
		}
		rows[i] = table.Row(paddedRow)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57"))

	t.SetStyles(s)

	// Print title if provided
	if opts.Title != "" {
		fmt.Println(opts.Title)
	}

	// Render table
	fmt.Println(t.View())

	return nil
}

// FormatMapToRows converts a map to a slice of string slices for table rendering
func FormatMapToRows(data map[string]string) []string {
	var rows []string
	for k, v := range data {
		rows = append(rows, fmt.Sprintf("%s: %s", k, v))
	}
	return rows
}

// JoinRows joins multiple rows into a single string
func JoinRows(rows []string) string {
	return strings.Join(rows, "\n")
}
