package output

import (
	"fmt"

	"github.com/Excoriate/aws-taggy/cli/internal/tui"
)

// This file provides additional rendering logic for validation results
// without duplicating the core rendering functions

// RenderDetailedTables provides an enhanced table-based rendering of validation results
func renderDetailedValidationTables(result *ValidationResult) error {
	// Validate input
	if result == nil {
		return fmt.Errorf("validation result cannot be nil")
	}

	// General Information Table
	generalInfo := [][]string{
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
			{Title: "Property", Width: 20},
			{Title: "Value", Width: 50, Flexible: true},
		},
		FlexibleColumns: true,
	}

	if err := tui.RenderTable(generalTableOpts, generalInfo); err != nil {
		return err
	}

	// Global Configuration Table
	globalConfig := [][]string{
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
			{Title: "Setting", Width: 25},
			{Title: "Value", Width: 45, Flexible: true},
		},
		FlexibleColumns: true,
	}

	if err := tui.RenderTable(globalTableOpts, globalConfig); err != nil {
		return err
	}

	// Resources Table
	resources := [][]string{}
	for _, service := range result.Resources.Services {
		resources = append(resources, []string{service, "Enabled"})
	}

	resourcesTableOpts := tui.TableOptions{
		Title: "🔍 Configured Resources",
		Columns: []tui.Column{
			{Title: "Service", Width: 20},
			{Title: "Status", Width: 10},
		},
		FlexibleColumns: true,
	}

	if err := tui.RenderTable(resourcesTableOpts, resources); err != nil {
		return err
	}

	// Display warnings if any
	if len(result.Warnings) > 0 {
		warnings := [][]string{}
		for _, w := range result.Warnings {
			warnings = append(warnings, []string{w})
		}

		warningsTableOpts := tui.TableOptions{
			Title: "⚠️  Warnings",
			Columns: []tui.Column{
				{Title: "Warning", Width: 70, Flexible: true},
			},
			FlexibleColumns: true,
		}

		if err := tui.RenderTable(warningsTableOpts, warnings); err != nil {
			return err
		}
	}

	return nil
}
