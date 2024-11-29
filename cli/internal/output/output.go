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
	ResourceID      string            `json:"resource_id" yaml:"resource_id"`
	ResourceType    string            `json:"resource_type" yaml:"resource_type"`
}

// Violation represents a specific tag compliance violation
type Violation struct {
	Type    string `json:"type" yaml:"type"`
	Message string `json:"message" yaml:"message"`
}

// ComplianceSummary provides an overview of compliance results
type ComplianceSummary struct {
	TotalResources        int                    `json:"total_resources" yaml:"total_resources"`
	CompliantResources    int                    `json:"compliant_resources" yaml:"compliant_resources"`
	NonCompliantResources int                    `json:"non_compliant_resources" yaml:"non_compliant_resources"`
	GlobalViolations      map[string]int         `json:"global_violations,omitempty" yaml:"global_violations,omitempty"`
	RuleResults           map[string]*RuleResult `json:"rule_results,omitempty" yaml:"rule_results,omitempty"`
}

// RuleResult represents the result of a specific compliance rule
type RuleResult struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Passed      bool   `json:"passed" yaml:"passed"`
	Failures    int    `json:"failures" yaml:"failures"`
}

// PlannedChecks represents the compliance checks that will be executed
type PlannedChecks struct {
	Rules []ComplianceRule `json:"rules" yaml:"rules"`
}

// ComplianceRule represents a single compliance rule to be checked
type ComplianceRule struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
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

// PrintConfigValidation prints a success message for configuration validation
func PrintConfigValidation() {
	fmt.Printf("\n✅ Configuration validation successful\n\n")
}

// PrintPlannedChecks prints the compliance checks that will be executed
func PrintPlannedChecks(checks PlannedChecks) {
	fmt.Printf("🔍 Planned compliance checks:\n\n")
	for _, rule := range checks.Rules {
		fmt.Printf("  • %s\n    %s\n", rule.Name, rule.Description)
	}
	fmt.Printf("\n")
}

// PrintComplianceSummary prints a detailed summary of the compliance results
func PrintComplianceSummary(summary ComplianceSummary) {
	fmt.Printf("\n📊 Compliance Summary:\n\n")
	fmt.Printf("Total Resources: %d\n", summary.TotalResources)
	fmt.Printf("Compliant: %d\n", summary.CompliantResources)
	fmt.Printf("Non-Compliant: %d\n\n", summary.NonCompliantResources)

	if len(summary.RuleResults) > 0 {
		fmt.Printf("Rule Results:\n")
		for _, result := range summary.RuleResults {
			status := "✅"
			if !result.Passed {
				status = "❌"
			}
			fmt.Printf("%s %s\n", status, result.Name)
			fmt.Printf("   Description: %s\n", result.Description)
			if !result.Passed {
				fmt.Printf("   Failures: %d\n", result.Failures)
			}
			fmt.Printf("\n")
		}
	}

	if len(summary.GlobalViolations) > 0 {
		fmt.Printf("Violation Types:\n")
		for vType, count := range summary.GlobalViolations {
			fmt.Printf("  🚨 %s: %d occurrences\n", vType, count)
		}
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
