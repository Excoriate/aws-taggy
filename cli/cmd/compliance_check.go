package cmd

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/cli/internal/output"
	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/compliance"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
)

// CheckCmd represents the compliance check command
type CheckCmd struct {
	Config    string `help:"Path to the tag compliance configuration file" required:"true"`
	Output    string `help:"Output format (table|json|yaml)" default:"table" enum:"table,json,yaml"`
	Table     bool   `help:"Display detailed information in tables" default:"false"`
	Clipboard bool   `help:"Copy output to clipboard" default:"false"`
}

// Run validates the configuration file and performs compliance checks
func (c *CheckCmd) Run() error {
	logger := o11y.DefaultLogger()
	logger.Info(fmt.Sprintf("🔍 Checking compliance configuration file: %s", c.Config))

	// Initialize configuration loader and validator
	loader := configuration.NewTaggyScanConfigLoader()
	fileValidator, err := configuration.NewConfigFileValidator(c.Config)
	if err != nil {
		return fmt.Errorf("failed to initialize file validator: %w", err)
	}

	// Validate file first
	if err := fileValidator.Validate(); err != nil {
		return fmt.Errorf("file validation failed: %w", err)
	}

	// Load configuration
	cfg, err := loader.LoadConfig(c.Config)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize config validator
	configValidator, err := configuration.NewConfigValidator(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize config validator: %w", err)
	}

	// Perform configuration validation
	if err := configValidator.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create compliance validator
	complianceValidator := compliance.NewTagValidator(cfg)

	// Prepare example test tags for demonstration
	testTags := []map[string]string{
		{
			"Environment": "production",
			"Owner":       "team@company.com",
		},
		{
			"Environment": "INVALID-ENV",
			"Owner":       "invalid-email",
		},
		{
			// Missing required tags
		},
	}

	// Validate tags and collect results
	var complianceResults []*compliance.ComplianceResult
	for _, tags := range testTags {
		result := complianceValidator.ValidateTags(tags)

		complianceResults = append(complianceResults, result)
	}

	// Generate compliance summary
	summary := compliance.GenerateSummary(complianceResults)

	// Prepare validation result for output
	result := output.ValidationResult{
		File:    c.Config,
		Valid:   true,
		Status:  "valid",
		Version: cfg.Version,
	}

	// Handle clipboard if requested
	if c.Clipboard {
		if err := output.WriteToClipboard(result); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✅ Compliance check result copied to clipboard!")
		return nil
	}

	// Create output formatter
	formatter := output.NewFormatter(c.Output)

	if formatter.IsStructured() {
		return formatter.Output(result)
	}

	// If table view is requested
	if c.Table {
		// Prepare table data
		tableData := [][]string{}
		for _, compResult := range complianceResults {
			tagsStr := formatTags(compResult.ResourceTags)
			complianceStatus := "✅ Compliant"
			if !compResult.IsCompliant {
				complianceStatus = "❌ Non-Compliant"
			}

			violationsStr := formatViolations(compResult.Violations)
			tableData = append(tableData, []string{tagsStr, complianceStatus, violationsStr})
		}

		// Add summary row
		tableData = append(tableData, []string{
			"Summary",
			fmt.Sprintf("Total: %d", summary.TotalResources),
			fmt.Sprintf("Compliant: %d, Non-Compliant: %d", summary.CompliantResources, summary.NonCompliantResources),
		})

		// Render table
		tableOpts := tui.TableOptions{
			Title: "Compliance Check Results",
			Columns: []tui.Column{
				{Title: "Resource Tags", Width: 40, Flexible: true},
				{Title: "Status", Width: 20},
				{Title: "Details", Width: 40, Flexible: true},
			},
			AutoWidth: true,
		}

		return tui.RenderTable(tableOpts, tableData)
	}

	// Default console output
	return output.RenderDefaultOutput(&result)
}

// Helper functions
func formatTags(tags map[string]string) string {
	if len(tags) == 0 {
		return "No Tags"
	}

	var result string
	for k, v := range tags {
		if result != "" {
			result += "\n"
		}
		result += fmt.Sprintf("%s: %s", k, v)
	}
	return result
}

func formatViolations(violations []compliance.Violation) string {
	if len(violations) == 0 {
		return "No Violations"
	}

	var result string
	for _, v := range violations {
		if result != "" {
			result += "\n"
		}
		result += fmt.Sprintf("%s: %s", v.Type, v.Message)
	}
	return result
}
