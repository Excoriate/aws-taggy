package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	Format(data interface{}) (string, error)
}

// JSONFormatter implements Formatter for JSON output
type JSONFormatter struct {
	Pretty bool
}

// Format formats the data as JSON
func (f *JSONFormatter) Format(data interface{}) (string, error) {
	var bytes []byte
	var err error

	if f.Pretty {
		bytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		bytes, err = json.Marshal(data)
	}

	if err != nil {
		return "", fmt.Errorf("failed to format as JSON: %w", err)
	}

	return string(bytes), nil
}

// TableFormatter implements Formatter for table output
type TableFormatter struct {
	Headers     []string
	FormatStyle string
}

// Format formats the data as a table
func (f *TableFormatter) Format(data interface{}) (string, error) {
	// Implementation depends on your specific table formatting needs
	// This is a basic example
	rows, ok := data.([][]string)
	if !ok {
		return "", fmt.Errorf("data must be [][]string for table formatting")
	}

	var sb strings.Builder

	// Write headers
	if len(f.Headers) > 0 {
		sb.WriteString(strings.Join(f.Headers, "\t") + "\n")
		sb.WriteString(strings.Repeat("-", 80) + "\n")
	}

	// Write rows
	for _, row := range rows {
		sb.WriteString(strings.Join(row, "\t") + "\n")
	}

	return sb.String(), nil
}

// NewJSONFormatter creates a new JSONFormatter
func NewJSONFormatter(pretty bool) Formatter {
	return &JSONFormatter{Pretty: pretty}
}

// NewTableFormatter creates a new TableFormatter
func NewTableFormatter(headers []string) Formatter {
	return &TableFormatter{Headers: headers}
}
