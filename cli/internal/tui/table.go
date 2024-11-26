package tui

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Column defines a table column with its properties
type Column struct {
	Title    string
	Key      string
	Width    int
	Flexible bool
}

// TableStyle defines the styling for the TUI table
type TableStyle struct {
	BaseStyle   lipgloss.Style
	HeaderStyle lipgloss.Style
	RowStyle    lipgloss.Style
	SelectedRow lipgloss.Style
}

// TableOptions contains configuration options for the table
type TableOptions struct {
	Title           string
	Columns         []Column
	Style           *TableStyle
	MaxHeight       int
	FlexibleColumns bool
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

// TableModel represents the state of a TUI table
type TableModel struct {
	table   table.Model
	style   TableStyle
	title   string
	columns []Column
	data    interface{}
	err     error
}

// NewTableModel creates a new TUI table model with flexible data handling
func NewTableModel(opts TableOptions, data interface{}) TableModel {
	// Use default style if not provided
	tableStyle := DefaultTableStyle()
	if opts.Style != nil {
		tableStyle = *opts.Style
	}

	// Set default max height if not provided
	if opts.MaxHeight == 0 {
		opts.MaxHeight = 10
	}

	// Convert data to rows first to analyze content width
	rows := extractRows(opts.Columns, data)

	// Calculate column widths if flexible columns are enabled
	tableColumns := make([]table.Column, len(opts.Columns))
	for i, col := range opts.Columns {
		width := col.Width

		if opts.FlexibleColumns && col.Flexible {
			// Calculate maximum content width for this column
			maxWidth := len(col.Title) // Start with header width
			for _, row := range rows {
				if len(row[i]) > maxWidth {
					maxWidth = len(row[i])
				}
			}
			// Add padding and ensure minimum width
			width = maxWidth + 2 // Add some padding
			if width < col.Width {
				width = col.Width // Use specified width as minimum
			}
		}

		if width == 0 {
			width = 20 // Default width
		}

		tableColumns[i] = table.Column{
			Title: col.Title,
			Width: width,
		}
	}

	// Create table with calculated widths
	t := table.New(
		table.WithColumns(tableColumns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(opts.MaxHeight),
	)

	t.SetStyles(table.Styles{
		Header:   tableStyle.HeaderStyle,
		Cell:     tableStyle.RowStyle,
		Selected: tableStyle.SelectedRow,
	})

	return TableModel{
		table:   t,
		style:   tableStyle,
		title:   opts.Title,
		columns: opts.Columns,
		data:    data,
	}
}

// extractRows converts the data into table rows based on column definitions
func extractRows(columns []Column, data interface{}) []table.Row {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var rows []table.Row
	switch val.Kind() {
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i)
			row := extractRow(columns, item)
			rows = append(rows, row)
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			item := val.MapIndex(key)
			row := extractRow(columns, item)
			rows = append(rows, row)
		}
	}

	return rows
}

// extractRow creates a single row from an item based on column definitions
func extractRow(columns []Column, item reflect.Value) table.Row {
	row := make(table.Row, len(columns))
	for i, col := range columns {
		value := extractValue(item, col.Key)
		row[i] = fmt.Sprintf("%v", value)
	}
	return row
}

// extractValue gets a value from an item using the column key
func extractValue(item reflect.Value, key string) interface{} {
	if item.Kind() == reflect.Ptr {
		item = item.Elem()
	}

	switch item.Kind() {
	case reflect.Struct:
		field := item.FieldByName(key)
		if field.IsValid() {
			return field.Interface()
		}
	case reflect.Map:
		value := item.MapIndex(reflect.ValueOf(key))
		if value.IsValid() {
			return value.Interface()
		}
	}
	return ""
}

// Init implements tea.Model
func (m TableModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m TableModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		Padding(0, 1)

	return fmt.Sprintf("%s\n%s",
		titleStyle.Render(m.title),
		m.style.BaseStyle.Render(m.table.View()))
}

// Render runs the table UI
func (m TableModel) Render() error {
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
