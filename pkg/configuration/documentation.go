package configuration

// DefaultDocumentation returns the default documentation content for AWS Taggy configuration
func DefaultDocumentation() string {
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
See the generated YAML file for a complete example with all available options.`
}

// GenerateDocumentationFilename generates a documentation filename from a configuration filename
func GenerateDocumentationFilename(configFile string) string {
	return configFile + ".md"
}
