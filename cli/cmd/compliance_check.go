package cmd

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/cli/internal/output"
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

// Run validates the configuration file and prepares for compliance checking
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
	validator, err := configuration.NewConfigValidator(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize config validator: %w", err)
	}

	// Prepare validation result
	result := output.ValidationResult{
		File:    c.Config,
		Valid:   true,
		Status:  "valid",
		Version: cfg.Version,
	}

	// Perform validation
	if err := validator.Validate(); err != nil {
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
	if c.Clipboard {
		if err := output.CopyToClipboard(result); err != nil {
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
		return output.RenderDetailedTables(result)
	}

	// Default console output
	return output.RenderDefaultOutput(result)
}
