package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"gopkg.in/yaml.v3"
)

type ConfigurationWriter struct {
	Config *configuration.TaggyScanConfig
	File   string
}

type DocumentationWriter struct {
	File string
}

func NewConfigurationWriter() *ConfigurationWriter {
	return &ConfigurationWriter{
		Config: &configuration.TaggyScanConfig{},
	}
}

func NewDocumentationWriter() *DocumentationWriter {
	return &DocumentationWriter{
		File: "how-to-customize-aws-taggy-configuration.md",
	}
}

func (w *ConfigurationWriter) WriteConfiguration(file string, overwrite bool) error {
	// Set default configuration
	w.SetDefaultConfig()

	// Validate the file path
	if file == "" {
		return fmt.Errorf("output file path cannot be empty")
	}

	// Check if file exists and overwrite is not allowed
	if !overwrite {
		if _, err := os.Stat(file); err == nil {
			return fmt.Errorf("configuration file already exists at %s. Use the -f flag to overwrite", file)
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory for configuration file: %w", err)
	}

	// Marshal the configuration to YAML
	yamlData, err := yaml.Marshal(w.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration to YAML: %w", err)
	}

	// Write the YAML to file
	if err := os.WriteFile(file, yamlData, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

func (w *ConfigurationWriter) SetDefaultConfig() {
	// Version of the configuration file
	// Helps with future compatibility and configuration management
	w.Config.Version = "1.0"

	// AWS Configuration
	// Define how AWS resources are discovered and processed
	w.Config.AWS.Regions.Mode = "all" // Scan all available regions by default
	batchSize := 20
	w.Config.AWS.BatchSize = &batchSize // Number of resources to process in parallel

	// Global Tagging Configuration
	// These settings apply across all resources unless overridden
	w.Config.Global.Enabled = true // Enable global tag compliance checks

	// Tag Criteria define the rules for tag validation
	w.Config.Global.TagCriteria.MinimumRequiredTags = 3 // Minimum number of tags required
	w.Config.Global.TagCriteria.MaxTags = 50            // Maximum number of tags allowed

	// Required Tags: Tags that MUST be present on all resources
	w.Config.Global.TagCriteria.RequiredTags = []string{
		"Environment", // Identifies the deployment environment (dev, staging, prod)
		"Owner",       // Team or individual responsible for the resource
		"Project",     // Project or initiative the resource belongs to
	}

	// Forbidden Tags: Tags that are not allowed
	w.Config.Global.TagCriteria.ForbiddenTags = []string{
		"Temporary", // Prevents resources with temporary tags from being considered compliant
		"Test",      // Excludes test resources from standard compliance
	}

	// Specific Tags: Exact tag key-value pairs that must be present
	w.Config.Global.TagCriteria.SpecificTags = map[string]string{
		"ComplianceLevel": "high",      // Indicates a high level of tag compliance
		"ManagedBy":       "terraform", // Identifies resources managed by Terraform
	}

	// Compliance Level: Overall tag compliance standard
	w.Config.Global.TagCriteria.ComplianceLevel = "high"

	// Resource-Specific Configurations
	// Define unique tagging rules for different AWS resource types
	w.Config.Resources = map[string]configuration.ResourceConfig{
		// S3 Bucket Specific Configuration
		"s3": {
			Enabled: true, // Enable tag compliance checks for S3 buckets
			TagCriteria: configuration.TagCriteria{
				MinimumRequiredTags: 4, // More strict tag requirements for S3
				RequiredTags: []string{
					"DataClassification", // Sensitivity level of data stored
					"BackupPolicy",       // Backup and retention strategy
					"Environment",        // Deployment environment
					"Owner",              // Resource ownership
				},
				ForbiddenTags: []string{
					"Temporary", // Prevent temporary buckets
					"Test",      // Exclude test buckets
				},
				SpecificTags: map[string]string{
					"EncryptionRequired": "true", // Enforce encryption for sensitive buckets
				},
				ComplianceLevel: "high", // Strict compliance for S3
			},
			// Exclude specific S3 buckets from compliance checks
			ExcludedResources: []configuration.ExcludedResource{
				{
					Pattern: "terraform-state-*", // Exclude Terraform state buckets
					Reason:  "Terraform state buckets managed separately",
				},
				{
					Pattern: "log-archive-*", // Exclude logging buckets
					Reason:  "Logging buckets excluded from standard compliance",
				},
			},
		},
		// EC2 Instance Specific Configuration
		"ec2": {
			Enabled: true, // Enable tag compliance checks for EC2 instances
			TagCriteria: configuration.TagCriteria{
				MinimumRequiredTags: 3, // Standard tag requirements
				RequiredTags: []string{
					"Application", // Application or service running on the instance
					"PatchGroup",  // Patch management group
					"Environment", // Deployment environment
				},
				ForbiddenTags: []string{
					"Temporary", // Prevent temporary instances
					"Test",      // Exclude test instances
				},
				SpecificTags: map[string]string{
					"AutoStop": "enabled", // Enable automatic instance stopping
				},
				ComplianceLevel: "standard", // Standard compliance level
			},
			// Exclude specific EC2 instances from compliance checks
			ExcludedResources: []configuration.ExcludedResource{
				{
					Pattern: "bastion-*", // Exclude bastion hosts
					Reason:  "Bastion hosts managed by security team",
				},
			},
		},
	}
}

func (w *DocumentationWriter) WriteDocumentation(configFile string) error {
	// Validate the file path
	if configFile == "" {
		return fmt.Errorf("configuration file path cannot be empty")
	}

	// Generate documentation file path by replacing the extension with .md
	docFile := strings.TrimSuffix(configFile, filepath.Ext(configFile)) + ".md"

	// Ensure directory exists
	dir := filepath.Dir(docFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory for documentation file: %w", err)
	}

	// Generate documentation content
	content := w.generateDocumentationContent()

	// Write the documentation to file
	if err := os.WriteFile(docFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write documentation file: %w", err)
	}

	return nil
}

func (w *DocumentationWriter) generateDocumentationContent() string {
	return `# AWS Taggy Configuration Guide

This document provides a comprehensive guide to configuring AWS Taggy for tag compliance management.

## Configuration Structure

### Version
The version field tracks the schema version of your configuration file, enabling future compatibility and schema evolution.

### AWS Configuration
#### Regions
- **mode**: Can be 'all' or 'specific'
  - 'all': Scans all supported AWS regions
  - 'specific': Only scans listed regions
- **list**: List of specific regions to scan (when mode is 'specific')

#### Batch Size
- **batch_size**: Controls the number of resources processed in parallel (default: 20)

### Global Settings
Global settings define the default tagging rules applied across all resources unless overridden.

#### Tag Criteria
- **minimum_required_tags**: Minimum number of tags required for compliance
- **max_tags**: Maximum number of tags allowed per resource
- **required_tags**: List of tags that must be present on every resource
- **forbidden_tags**: List of tags that are not allowed
- **specific_tags**: Exact tag key-value pairs that must be present
- **compliance_level**: Overall tag compliance standard (e.g., 'high', 'standard')

### Resource-Specific Configurations
Define custom tagging rules for different AWS resource types.

#### Example: S3 Configuration
- **enabled**: Enable/disable tag compliance for S3
- **tag_criteria**: Custom tag requirements for S3 buckets
  - **minimum_required_tags**: S3-specific minimum tag requirement
  - **required_tags**: S3-specific required tags
  - **forbidden_tags**: S3-specific forbidden tags
  - **specific_tags**: S3-specific required tag key-value pairs
  - **compliance_level**: S3-specific compliance level
- **excluded_resources**: Patterns for S3 buckets to exclude from compliance checks

### Compliance Levels
Define different compliance standards with specific requirements.

#### High Compliance
Strictest tagging requirements with comprehensive metadata and security validations.

#### Standard Compliance
Moderate tagging requirements for general resource management.

### Tag Validation Rules
Rules for validating tag keys and values.

#### Key Format Rules
- Patterns for valid tag keys
- Length restrictions
- Allowed prefixes and suffixes

#### Value Validation
- Allowed characters
- Disallowed values
- Length constraints
- Case sensitivity rules

### Notifications
Configure alerts and reports for non-compliant resources.

#### Slack Notifications
- Channel configurations for different priority levels
- Alert settings

#### Email Notifications
- Recipient configuration
- Reporting frequency settings

## Best Practices
1. Start with minimum required tags and gradually increase requirements
2. Use consistent naming conventions
3. Regularly review and update compliance rules
4. Document exceptions in excluded_resources
5. Implement strict validation for production environments

## Example Configuration
See the generated YAML file for a complete example with all available options.
`
}
