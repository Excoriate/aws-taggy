package cmd

import (
	"context"
	"fmt"

	"github.com/Excoriate/aws-taggy/cli/internal/output"
	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/compliance"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/inspector"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
	"github.com/Excoriate/aws-taggy/pkg/taggy"
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

	// Initialize taggy client
	client, err := taggy.New(c.Config)
	if err != nil {
		return fmt.Errorf("failed to initialize taggy client: %w", err)
	}

	// Initialize scanner manager
	scannerMgr, err := inspector.NewScannerManager(*client.Config())
	if err != nil {
		return fmt.Errorf("failed to create scanner manager: %w", err)
	}

	// Scan resources
	logger.Info("🔍 Scanning AWS resources...")
	ctx := context.Background()
	if err := scannerMgr.Scan(ctx); err != nil {
		return fmt.Errorf("failed to scan resources: %w", err)
	}

	// Get scan results
	scanResults := scannerMgr.GetResults()

	// Create compliance validator
	complianceValidator := compliance.NewTagValidator(cfg)

	// Validate tags and collect results
	var complianceResults []*output.ComplianceResult
	for _, result := range scanResults {
		for _, resource := range result.Resources {
			validationResult := complianceValidator.ValidateTags(resource.Tags)

			// Convert compliance.ComplianceResult to output.ComplianceResult
			outputResult := &output.ComplianceResult{
				IsCompliant:     validationResult.IsCompliant,
				ResourceTags:    validationResult.ResourceTags,
				ComplianceLevel: string(validationResult.ComplianceLevel),
				ResourceID:      resource.ID,
				ResourceType:    resource.Type,
			}

			// Convert violations
			for _, v := range validationResult.Violations {
				outputResult.Violations = append(outputResult.Violations, output.Violation{
					Type:    string(v.Type),
					Message: v.Message,
				})
			}

			complianceResults = append(complianceResults, outputResult)
		}
	}

	// Convert output results back to compliance results for summary generation
	var internalResults []*compliance.ComplianceResult
	for _, result := range complianceResults {
		internalResult := &compliance.ComplianceResult{
			IsCompliant:     result.IsCompliant,
			ResourceTags:    result.ResourceTags,
			ComplianceLevel: compliance.ComplianceLevel(result.ComplianceLevel),
		}

		// Convert violations
		for _, v := range result.Violations {
			internalResult.Violations = append(internalResult.Violations, compliance.Violation{
				Type:    compliance.ViolationType(v.Type),
				Message: v.Message,
			})
		}

		internalResults = append(internalResults, internalResult)
	}

	// Generate compliance summary
	summary := compliance.GenerateSummary(internalResults)

	// Prepare validation result for output
	result := output.ValidationResult{
		File:              c.Config,
		Valid:             true,
		Status:            "valid",
		Version:           cfg.Version,
		ComplianceResults: complianceResults,
		ComplianceSummary: &output.ComplianceSummary{
			TotalResources:        summary.TotalResources,
			CompliantResources:    summary.CompliantResources,
			NonCompliantResources: summary.NonCompliantResources,
			GlobalViolations:      make(map[string]int),
		},
	}

	// Convert global violations
	for vType, count := range summary.GlobalViolations {
		result.ComplianceSummary.GlobalViolations[string(vType)] = count
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
			resourceInfo := fmt.Sprintf("%s (%s)", compResult.ResourceID, compResult.ResourceType)
			tagsStr := formatTags(compResult.ResourceTags)
			complianceStatus := "✅ Compliant"
			if !compResult.IsCompliant {
				complianceStatus = "❌ Non-Compliant"
			}

			violationsStr := formatViolations(compResult.Violations)
			tableData = append(tableData, []string{resourceInfo, tagsStr, complianceStatus, violationsStr})
		}

		// Add summary row
		tableData = append(tableData, []string{
			"Summary",
			fmt.Sprintf("Total: %d", summary.TotalResources),
			fmt.Sprintf("Compliant: %d", summary.CompliantResources),
			fmt.Sprintf("Non-Compliant: %d", summary.NonCompliantResources),
		})

		// Render table
		tableOpts := tui.TableOptions{
			Title: "Compliance Check Results",
			Columns: []tui.Column{
				{Title: "Resource", Width: 30, Flexible: true},
				{Title: "Tags", Width: 40, Flexible: true},
				{Title: "Status", Width: 20},
				{Title: "Violations", Width: 40, Flexible: true},
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

func formatViolations(violations []output.Violation) string {
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
