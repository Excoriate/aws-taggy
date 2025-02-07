package cmd

import (
	"fmt"
	"strings"

	"github.com/Excoriate/aws-taggy/cli/internal/output"
	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
)

// ValidateCmd represents the validate subcommand
type ValidateCmd struct {
	Config    string `help:"Path to the tag validation configuration file" required:"true"`
	Output    string `help:"Output format (table|json|yaml)" default:"table" enum:"table,json,yaml,TABLE,JSON,YAML"`
	Table     bool   `help:"Display detailed information in tables" default:"false"`
	Clipboard bool   `help:"Copy output to clipboard" default:"false"`
}

// Run method for ValidateCmd implements the configuration validation logic
func (v *ValidateCmd) Run() error {
	logger := o11y.DefaultLogger()
	logger.Info(fmt.Sprintf("🔍 Validating configuration file: %s", v.Config))

	// Initialize configuration loader and validator
	loader := configuration.NewTaggyScanConfigLoader()

	// Load configuration
	cfg, err := loader.LoadConfig(v.Config)
	if err != nil {
		return fmt.Errorf("failed to load configuration file %s: %w", v.Config, err)
	}

	// Initialize config validator
	validator, err := configuration.NewContentValidator(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize configuration validator for file %s: %w", v.Config, err)
	}

	// Prepare validation result
	result := output.ValidationResult{
		File:    v.Config,
		Valid:   true,
		Status:  "valid",
		Version: cfg.Version,
	}

	// Perform validation
	if err := validator.ValidateContent(); err != nil {
		result.Valid = false
		result.Status = "invalid"
		result.Errors = append(result.Errors, err.Error())
	}

	// Collect resource statistics and global config
	result.GlobalConfig.Enabled = cfg.Global.Enabled
	result.GlobalConfig.MinRequiredTags = cfg.Global.TagCriteria.MinimumRequiredTags
	result.GlobalConfig.RequiredTags = cfg.Global.TagCriteria.RequiredTags
	result.GlobalConfig.ForbiddenTags = cfg.Global.TagCriteria.ForbiddenTags
	result.GlobalConfig.ComplianceLevel = cfg.Global.TagCriteria.ComplianceLevel
	if cfg.AWS.BatchSize != nil {
		result.GlobalConfig.BatchSize = *cfg.AWS.BatchSize
	} else {
		result.GlobalConfig.BatchSize = 20 // Default batch size
	}
	result.GlobalConfig.NotificationsSetup = cfg.Notifications.Slack.Enabled || cfg.Notifications.Email.Enabled

	// Collect compliance levels
	for level := range cfg.ComplianceLevels {
		result.ComplianceLevels = append(result.ComplianceLevels, level)
	}

	// Collect resource information
	for resourceType, resourceConfig := range cfg.Resources {
		result.Resources.Total++
		if resourceConfig.Enabled {
			result.Resources.Enabled++
			result.Resources.Services = append(result.Resources.Services, resourceType)
		}
	}

	// Add warnings for potential issues
	if result.Resources.Enabled == 0 {
		result.Warnings = append(result.Warnings, "No resources are enabled for scanning")
	}
	if !result.GlobalConfig.NotificationsSetup {
		result.Warnings = append(result.Warnings, "No notification channels are configured")
	}

	// Handle clipboard if requested
	if v.Clipboard {
		if err := output.WriteToClipboard(result); err != nil {
			return fmt.Errorf("failed to copy validation result to clipboard for file %s: %w", v.Config, err)
		}
		fmt.Println("✅ Validation result copied to clipboard!")
		return nil
	}

	// Create output formatter
	formatter := output.NewFormatter(v.Output)

	if formatter.IsStructured() {
		if err := formatter.Output(result); err != nil {
			return fmt.Errorf("failed to output structured validation result for file %s: %w", v.Config, err)
		}
		return nil
	}

	// If table view is requested
	if v.Table {
		// Prepare table data
		tableData := [][]string{
			{"Configuration File", v.Config},
			{"Validation Status", result.Status},
			{"Version", result.Version},
			{"Compliance Levels", strings.Join(result.ComplianceLevels, ", ")},
		}

		// Add resources information
		tableData = append(tableData,
			[]string{"Total Resources", fmt.Sprintf("%d", result.Resources.Total)},
			[]string{"Enabled Resources", fmt.Sprintf("%d", result.Resources.Enabled)},
			[]string{"Enabled Services", strings.Join(result.Resources.Services, ", ")},
		)

		// Add global config information
		tableData = append(tableData,
			[]string{"Minimum Required Tags", fmt.Sprintf("%d", result.GlobalConfig.MinRequiredTags)},
			[]string{"Required Tags", strings.Join(result.GlobalConfig.RequiredTags, ", ")},
			[]string{"Compliance Level", result.GlobalConfig.ComplianceLevel},
			[]string{"Batch Size", fmt.Sprintf("%d", result.GlobalConfig.BatchSize)},
		)

		// Add warnings if any
		if len(result.Warnings) > 0 {
			tableData = append(tableData,
				[]string{"Warnings", strings.Join(result.Warnings, "\n")},
			)
		}

		// Render table
		tableOpts := tui.TableOptions{
			Title: "Configuration Validation Results",
			Columns: []tui.Column{
				{Title: "Property", Width: 30, Flexible: true},
				{Title: "Value", Width: 50, Flexible: true},
			},
			AutoWidth: true,
		}

		if err := tui.RenderTable(tableOpts, tableData); err != nil {
			return fmt.Errorf("failed to render validation results table for file %s: %w", v.Config, err)
		}
		return nil
	}

	// Default console output
	if err := output.RenderDefaultOutput(&result); err != nil {
		return fmt.Errorf("failed to render default validation output for file %s: %w", v.Config, err)
	}

	return nil
}
