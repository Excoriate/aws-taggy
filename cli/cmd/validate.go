package cmd

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/cli/internal/output"
	"github.com/Excoriate/aws-taggy/cli/internal/tui"
	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/Excoriate/aws-taggy/pkg/o11y"
)

// ValidateCmd represents the validate subcommand
type ValidateCmd struct {
	Config    string `help:"Path to the tag validation configuration file" required:"true"`
	Output    string `help:"Output format (table|json|yaml)" default:"table" enum:"table,json,yaml"`
	Table     bool   `help:"Display detailed information in tables" default:"false"`
	Clipboard bool   `help:"Copy output to clipboard" default:"false"`
}

// ValidationResult represents the structured output of validation
type ValidationResult struct {
	Status    string   `json:"status" yaml:"status"`
	File      string   `json:"file" yaml:"file"`
	Valid     bool     `json:"valid" yaml:"valid"`
	Errors    []string `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings  []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Version   string   `json:"version" yaml:"version"`
	Resources struct {
		Total    int      `json:"total" yaml:"total"`
		Enabled  int      `json:"enabled" yaml:"enabled"`
		Services []string `json:"services" yaml:"services"`
	} `json:"resources" yaml:"resources"`
	GlobalConfig struct {
		Enabled            bool     `json:"enabled" yaml:"enabled"`
		MinRequiredTags    int      `json:"min_required_tags" yaml:"min_required_tags"`
		RequiredTags       []string `json:"required_tags" yaml:"required_tags"`
		ForbiddenTags      []string `json:"forbidden_tags" yaml:"forbidden_tags"`
		ComplianceLevel    string   `json:"compliance_level" yaml:"compliance_level"`
		BatchSize          int      `json:"batch_size" yaml:"batch_size"`
		NotificationsSetup bool     `json:"notifications_setup" yaml:"notifications_setup"`
	} `json:"global_config" yaml:"global_config"`
	ComplianceLevels []string `json:"compliance_levels" yaml:"compliance_levels"`
}

// Run method for ValidateCmd implements the configuration validation logic
func (v *ValidateCmd) Run() error {
	logger := o11y.DefaultLogger()
	logger.Info(fmt.Sprintf("🔍 Validating configuration file: %s", v.Config))

	// Initialize configuration loader and validator
	loader := configuration.NewTaggyScanConfigLoader()
	fileValidator, err := configuration.NewConfigFileValidator(v.Config)
	if err != nil {
		return fmt.Errorf("failed to initialize file validator: %w", err)
	}

	// Validate file first
	if err := fileValidator.Validate(); err != nil {
		return fmt.Errorf("file validation failed: %w", err)
	}

	// Load configuration
	cfg, err := loader.LoadConfig(v.Config)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize config validator
	validator, err := configuration.NewConfigValidator(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize config validator: %w", err)
	}

	// Prepare validation result
	result := ValidationResult{
		File:    v.Config,
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
	if v.Clipboard {
		if err := output.CopyToClipboard(result); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✅ Validation result copied to clipboard!")
		return nil
	}

	// Create output formatter
	formatter := output.NewFormatter(v.Output)

	if formatter.IsStructured() {
		return formatter.Output(result)
	}

	// If table view is requested
	if v.Table {
		return renderDetailedTables(result)
	}

	// Default console output
	return renderDefaultOutput(result)
}

func renderDetailedTables(result ValidationResult) error {
	// General Information Table
	generalInfo := []struct {
		Property string
		Value    string
	}{
		{"Status", result.Status},
		{"Configuration File", result.File},
		{"Version", result.Version},
		{"Valid", fmt.Sprintf("%v", result.Valid)},
		{"Total Resources", fmt.Sprintf("%d", result.Resources.Total)},
		{"Enabled Resources", fmt.Sprintf("%d", result.Resources.Enabled)},
	}

	generalTableOpts := tui.TableOptions{
		Title: "📋 General Configuration Information",
		Columns: []tui.Column{
			{Title: "Property", Key: "Property", Width: 20},
			{Title: "Value", Key: "Value", Width: 50, Flexible: true},
		},
		FlexibleColumns: true,
	}

	if err := tui.NewTableModel(generalTableOpts, generalInfo).Render(); err != nil {
		return err
	}

	// Global Configuration Table
	globalConfig := []struct {
		Setting string
		Value   string
	}{
		{"Enabled", fmt.Sprintf("%v", result.GlobalConfig.Enabled)},
		{"Minimum Required Tags", fmt.Sprintf("%d", result.GlobalConfig.MinRequiredTags)},
		{"Required Tags", fmt.Sprintf("%v", result.GlobalConfig.RequiredTags)},
		{"Forbidden Tags", fmt.Sprintf("%v", result.GlobalConfig.ForbiddenTags)},
		{"Compliance Level", result.GlobalConfig.ComplianceLevel},
		{"Batch Size", fmt.Sprintf("%d", result.GlobalConfig.BatchSize)},
		{"Notifications Setup", fmt.Sprintf("%v", result.GlobalConfig.NotificationsSetup)},
	}

	globalTableOpts := tui.TableOptions{
		Title: "🌍 Global Tag Configuration",
		Columns: []tui.Column{
			{Title: "Setting", Key: "Setting", Width: 25},
			{Title: "Value", Key: "Value", Width: 45, Flexible: true},
		},
		FlexibleColumns: true,
	}

	if err := tui.NewTableModel(globalTableOpts, globalConfig).Render(); err != nil {
		return err
	}

	// Resources Table
	resources := []struct {
		Service string
		Status  string
	}{}

	for _, service := range result.Resources.Services {
		resources = append(resources, struct {
			Service string
			Status  string
		}{
			Service: service,
			Status:  "Enabled",
		})
	}

	resourcesTableOpts := tui.TableOptions{
		Title: "🔍 Configured Resources",
		Columns: []tui.Column{
			{Title: "Service", Key: "Service", Width: 20},
			{Title: "Status", Key: "Status", Width: 10},
		},
		FlexibleColumns: true,
	}

	if err := tui.NewTableModel(resourcesTableOpts, resources).Render(); err != nil {
		return err
	}

	// Display warnings if any
	if len(result.Warnings) > 0 {
		warnings := []struct {
			Warning string
		}{}
		for _, w := range result.Warnings {
			warnings = append(warnings, struct{ Warning string }{Warning: w})
		}

		warningsTableOpts := tui.TableOptions{
			Title: "⚠️  Warnings",
			Columns: []tui.Column{
				{Title: "Warning", Key: "Warning", Width: 70, Flexible: true},
			},
			FlexibleColumns: true,
		}

		if err := tui.NewTableModel(warningsTableOpts, warnings).Render(); err != nil {
			return err
		}
	}

	return nil
}

func renderDefaultOutput(result ValidationResult) error {
	if !result.Valid {
		fmt.Printf("❌ Configuration validation failed for %s\n\n", result.File)
		fmt.Println("Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("configuration is invalid")
	}

	fmt.Printf("✅ Configuration file %s is valid\n\n", result.File)
	fmt.Printf("Version: %s\n\n", result.Version)

	fmt.Println("Global Configuration:")
	fmt.Printf("  Enabled: %v\n", result.GlobalConfig.Enabled)
	fmt.Printf("  Minimum Required Tags: %d\n", result.GlobalConfig.MinRequiredTags)
	fmt.Printf("  Required Tags: %v\n", result.GlobalConfig.RequiredTags)
	fmt.Printf("  Forbidden Tags: %v\n", result.GlobalConfig.ForbiddenTags)
	fmt.Printf("  Compliance Level: %s\n", result.GlobalConfig.ComplianceLevel)
	fmt.Printf("  Batch Size: %d\n", result.GlobalConfig.BatchSize)
	fmt.Printf("  Notifications Setup: %v\n\n", result.GlobalConfig.NotificationsSetup)

	fmt.Println("Resource Summary:")
	fmt.Printf("  Total Resources: %d\n", result.Resources.Total)
	fmt.Printf("  Enabled Resources: %d\n", result.Resources.Enabled)
	fmt.Printf("  Services: %v\n\n", result.Resources.Services)

	fmt.Printf("Compliance Levels: %v\n\n", result.ComplianceLevels)

	if len(result.Warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	return nil
}
