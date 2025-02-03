package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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
	Config     string `help:"Path to the tag compliance configuration file" required:"true"`
	Output     string `help:"Output format (table|json|yaml)" default:"table" enum:"table,json,yaml"`
	Table      bool   `help:"Display detailed information in tables" default:"false"`
	Detailed   bool   `help:"Show detailed compliance results for each resource" default:"false"`
	Clipboard  bool   `help:"Copy output to clipboard" default:"false"`
	OutputFile string `help:"Write detailed JSON output to specified file" type:"path"`
	Resource   string `help:"Filter compliance check for a specific resource (name or ARN)" optional:"true"`
}

// DetailedComplianceResult represents a detailed view of compliance results
type DetailedComplianceResult struct {
	Summary         output.ComplianceSummary      `json:"summary"`
	ResourceResults []*output.ComplianceResult    `json:"resource_results"`
	ValidationRules map[string]*output.RuleResult `json:"validation_rules"`
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
		return fmt.Errorf("failed to load configuration from file %s: %w. Please check the configuration file path and its contents", c.Config, err)
	}

	// Initialize config validator
	configValidator, err := configuration.NewContentValidator(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize configuration validator for file %s: %w. Ensure the configuration is valid and follows the expected schema", c.Config, err)
	}

	// Perform configuration validation
	if err := configValidator.ValidateContent(); err != nil {
		return fmt.Errorf("configuration validation failed for file %s: %w. Review the configuration and ensure all required fields are correctly specified", c.Config, err)
	}

	// Print configuration validation success
	output.PrintConfigValidation()

	// Print planned compliance checks
	plannedChecks := output.PlannedChecks{
		Rules: []output.ComplianceRule{
			{
				Name:        "Required Tags",
				Description: "Validates that all required tags are present",
			},
			{
				Name:        "Tag Value Format",
				Description: "Ensures tag values match specified formats and patterns",
			},
			{
				Name:        "Allowed Values",
				Description: "Verifies tag values are within allowed sets",
			},
			{
				Name:        "Case Sensitivity",
				Description: "Checks if tag keys and values follow case requirements",
			},
		},
	}
	output.PrintPlannedChecks(plannedChecks)

	// Initialize taggy client
	client, err := taggy.New(c.Config)
	if err != nil {
		return fmt.Errorf("failed to initialize taggy client with configuration %s: %w. Check the configuration and ensure all required parameters are set", c.Config, err)
	}

	// Initialize scanner manager
	inspectorMgr, err := inspector.NewInspectorManagerFromConfig(*client.Config())
	if err != nil {
		return fmt.Errorf("failed to create scanner manager from configuration: %w. Verify the AWS configuration and region settings", err)
	}

	// Scan resources
	logger.Info("🔍 Scanning AWS resources...")
	ctx := context.Background()
	if err := inspectorMgr.Inspect(ctx); err != nil {
		return fmt.Errorf("failed to scan AWS resources: %w. Check AWS credentials, permissions, and network connectivity", err)
	}

	// Get scan results
	inspectResults := inspectorMgr.GetResults()

	// Filter resources if Resource flag is provided
	if c.Resource != "" {
		logger.Info(fmt.Sprintf("🔍 Filtering resources matching: %s", c.Resource))
		filteredResults := make(map[string]*inspector.InspectResult)

		for resourceType, result := range inspectResults {
			filteredResources := make([]inspector.ResourceMetadata, 0)
			for _, resource := range result.Resources {
				// Check if resource matches by ID or ARN
				if resource.ID == c.Resource ||
					resource.Details.ARN == c.Resource ||
					resource.Details.Name == c.Resource {
					filteredResources = append(filteredResources, resource)
				}
			}

			if len(filteredResources) > 0 {
				filteredResult := &inspector.InspectResult{
					Resources:      filteredResources,
					StartTime:      result.StartTime,
					EndTime:        result.EndTime,
					Duration:       result.Duration,
					Region:         result.Region,
					TotalResources: len(filteredResources),
					Errors:         result.Errors,
				}
				filteredResults[resourceType] = filteredResult
			}
		}

		// If no resources match the filter, return an error
		if len(filteredResults) == 0 {
			return fmt.Errorf("no resources found matching the resource filter: %s", c.Resource)
		}

		// Safely get the number of filtered resources
		var totalFilteredResources int
		for _, result := range filteredResults {
			totalFilteredResources += len(result.Resources)
		}
		logger.Info(fmt.Sprintf("✅ Found %d resources matching the filter", totalFilteredResources))

		// Update inspectResults with filtered results
		inspectResults = filteredResults
	}

	// Create compliance validator
	complianceValidator := compliance.NewTagValidator(cfg)

	// Validate tags and collect results
	var complianceResults []*output.ComplianceResult
	ruleResults := make(map[string]*output.RuleResult)

	// Initialize rule results
	ruleResults["required_tags"] = &output.RuleResult{
		Name:        "Required Tags",
		Description: "Validates that all required tags are present",
		Passed:      true,
	}
	ruleResults["tag_format"] = &output.RuleResult{
		Name:        "Tag Value Format",
		Description: "Ensures tag values match specified formats and patterns",
		Passed:      true,
	}
	ruleResults["allowed_values"] = &output.RuleResult{
		Name:        "Allowed Values",
		Description: "Verifies tag values are within allowed sets",
		Passed:      true,
	}
	ruleResults["case_sensitivity"] = &output.RuleResult{
		Name:        "Case Sensitivity",
		Description: "Checks if tag keys and values follow case requirements",
		Passed:      true,
	}

	for _, result := range inspectResults {
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

			// Convert violations and update rule results
			for _, v := range validationResult.Violations {
				outputResult.Violations = append(outputResult.Violations, output.Violation{
					Type:    string(v.Type),
					Message: v.Message,
				})

				// Update rule results based on violation types
				switch v.Type {
				case "missing_required_tag":
					ruleResults["required_tags"].Passed = false
					ruleResults["required_tags"].Failures++
				case "invalid_format":
					ruleResults["tag_format"].Passed = false
					ruleResults["tag_format"].Failures++
				case "invalid_value":
					ruleResults["allowed_values"].Passed = false
					ruleResults["allowed_values"].Failures++
				case "case_mismatch":
					ruleResults["case_sensitivity"].Passed = false
					ruleResults["case_sensitivity"].Failures++
				}
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

	// Create final summary with rule results
	finalSummary := output.ComplianceSummary{
		TotalResources:        summary.TotalResources,
		CompliantResources:    summary.CompliantResources,
		NonCompliantResources: summary.NonCompliantResources,
		GlobalViolations:      make(map[string]int),
		RuleResults:           ruleResults,
	}

	// Convert global violations
	for vType, count := range summary.GlobalViolations {
		finalSummary.GlobalViolations[string(vType)] = count
	}

	// Create detailed compliance result
	detailedResult := &DetailedComplianceResult{
		ResourceResults: complianceResults,
		ValidationRules: ruleResults,
		Summary:         finalSummary,
	}

	// Handle JSON output to file if specified
	if c.OutputFile != "" {
		jsonData, err := json.MarshalIndent(detailedResult, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON data: %w", err)
		}
		if err := os.WriteFile(c.OutputFile, jsonData, 0o644); err != nil {
			return fmt.Errorf("failed to write JSON to file: %w", err)
		}
		logger.Info(fmt.Sprintf("✅ Detailed compliance results written to %s", c.OutputFile))
	}

	// Handle clipboard if requested
	if c.Clipboard {
		if err := output.WriteToClipboard(detailedResult); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✅ Compliance check result copied to clipboard!")
		return nil
	}

	// Create output formatter
	formatter := output.NewFormatter(c.Output)

	if formatter.IsStructured() {
		return formatter.Output(detailedResult)
	}

	// If table view is requested
	if c.Table {
		return renderDetailedTable(complianceResults, finalSummary)
	}

	// Print the compliance summary
	output.PrintComplianceSummary(finalSummary)

	// If detailed output is requested, print resource-specific results
	if c.Detailed {
		fmt.Printf("\n🔍 Detailed Resource Results:\n\n")
		for _, result := range complianceResults {
			status := "✅"
			if !result.IsCompliant {
				status = "❌"
			}
			fmt.Printf("%s Resource: %s (%s)\n", status, result.ResourceID, result.ResourceType)
			fmt.Printf("   Tags:\n")
			for k, v := range result.ResourceTags {
				fmt.Printf("      %s: %s\n", k, v)
			}
			if !result.IsCompliant {
				fmt.Printf("   Violations:\n")
				for _, v := range result.Violations {
					fmt.Printf("      • %s: %s\n", v.Type, v.Message)
				}
			}
			fmt.Printf("\n")
		}
	}

	return nil
}

func renderDetailedTable(results []*output.ComplianceResult, summary output.ComplianceSummary) error {
	// Prepare table data
	tableData := [][]string{}
	for _, compResult := range results {
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
