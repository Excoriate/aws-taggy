package output

import (
	"fmt"
)

// RenderDetailedTables renders detailed compliance results in a table format
func RenderDetailedTables(result *ValidationResult) error {
	if result == nil {
		return fmt.Errorf("validation result cannot be nil")
	}
	return renderDetailedValidationTables(result)
}

// RenderDefaultOutput renders a default console output for the validation result
func RenderDefaultOutput(result *ValidationResult) error {
	if result == nil {
		return fmt.Errorf("validation result cannot be nil")
	}

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
