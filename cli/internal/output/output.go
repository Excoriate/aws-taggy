package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ComplianceResult represents a single tag compliance validation result
type ComplianceResult struct {
	IsCompliant     bool              `json:"is_compliant" yaml:"is_compliant"`
	ResourceTags    map[string]string `json:"resource_tags" yaml:"resource_tags"`
	Violations      []Violation       `json:"violations,omitempty" yaml:"violations,omitempty"`
	ComplianceLevel string            `json:"compliance_level,omitempty" yaml:"compliance_level,omitempty"`
}

// Violation represents a specific tag compliance violation
type Violation struct {
	Type    string `json:"type" yaml:"type"`
	Message string `json:"message" yaml:"message"`
}

// ComplianceSummary provides an overview of compliance results
type ComplianceSummary struct {
	TotalResources        int            `json:"total_resources" yaml:"total_resources"`
	CompliantResources    int            `json:"compliant_resources" yaml:"compliant_resources"`
	NonCompliantResources int            `json:"non_compliant_resources" yaml:"non_compliant_resources"`
	GlobalViolations      map[string]int `json:"global_violations,omitempty" yaml:"global_violations,omitempty"`
}

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
		return outputJSON(data)
	case FormatYAML:
		return outputYAML(data)
	default:
		return fmt.Errorf("unsupported output format: %s", f.Format)
	}
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	return encoder.Encode(data)
}
