package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Format represents the supported output formats
type Format string

const (
	// FormatJSON represents JSON output format
	FormatJSON Format = "json"
	// FormatYAML represents YAML output format
	FormatYAML Format = "yaml"
	// FormatTable represents the default table output format
	FormatTable Format = "table"
)

// Formatter handles the output formatting for different formats
type Formatter struct {
	Format Format
}

// NewFormatter creates a new Formatter instance
func NewFormatter(format string) *Formatter {
	switch format {
	case string(FormatJSON):
		return &Formatter{Format: FormatJSON}
	case string(FormatYAML):
		return &Formatter{Format: FormatYAML}
	default:
		return &Formatter{Format: FormatTable}
	}
}

// IsStructured returns true if the format is JSON or YAML
func (f *Formatter) IsStructured() bool {
	return f.Format == FormatJSON || f.Format == FormatYAML
}

// Output formats and prints the data according to the specified format
func (f *Formatter) Output(data interface{}) error {
	switch f.Format {
	case FormatJSON:
		return f.outputJSON(data)
	case FormatYAML:
		return f.outputYAML(data)
	default:
		return fmt.Errorf("unsupported output format: %s", f.Format)
	}
}

func (f *Formatter) outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *Formatter) outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	return encoder.Encode(data)
}
