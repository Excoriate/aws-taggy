package output

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/cli/internal/tui"
)

// RenderDetailedTables renders a detailed table view of the validation result
func RenderDetailedTables(result ValidationResult) error {
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

// RenderDefaultOutput renders a default text-based output of the validation result
func RenderDefaultOutput(result ValidationResult) error {
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
