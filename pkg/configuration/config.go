package configuration

import (
	"regexp"
)

// TaggyScanConfig represents the overall configuration structure for the AWS tag management tool.
// It contains global settings, resource-specific configurations, compliance levels,
// tag validation rules, and notification settings for comprehensive AWS resource tag management.
//
// The configuration allows for fine-grained control over tag requirements,
// compliance levels, and notification mechanisms across different AWS resource types.
type TaggyScanConfig struct {
	// Version of the configuration file format
	Version string `yaml:"version"`

	// Global defines default configuration settings that apply across all resources
	Global GlobalConfig `yaml:"global"`

	// Resources contains configuration specific to individual resource types
	Resources map[string]ResourceConfig `yaml:"resources"`

	// ComplianceLevels defines different levels of tag compliance with their specific requirements
	ComplianceLevels map[string]ComplianceLevel `yaml:"compliance_levels"`

	// TagValidation contains rules for validating tags across resources
	TagValidation TagValidation `yaml:"tag_validation"`

	// Notifications manages the settings for reporting tag inspection results
	Notifications NotificationConfig `yaml:"notifications"`

	// AWS configuration for region scanning
	AWS AWSConfig `yaml:"aws"`
}

// GlobalConfig defines the default configuration settings that apply across all resources.
// It includes batch processing size, required and forbidden tags, and specific tag requirements.
type GlobalConfig struct {
	// Enabled determines if global configuration is active
	Enabled bool `yaml:"enabled"`

	// BatchSize specifies the default number of resources to process in a single batch
	// If not set, a system-default batch size will be used
	// This serves as a fallback/default for resource-specific and provider-specific batch sizes
	BatchSize *int `yaml:"batch_size,omitempty"`

	// TagCriteria defines the default tag validation rules for all resources
	TagCriteria TagCriteria `yaml:"tag_criteria"`
}

// ResourceConfig provides configuration specific to individual resource types.
// It allows for more granular control over tag requirements, exclusions, and processing.
type ResourceConfig struct {
	// Enabled determines if this resource type is subject to tag inspection
	Enabled bool `yaml:"enabled"`

	// Regions is an optional list of regions to scan for this specific resource type
	// If set, it overrides the global AWS regions configuration
	Regions []string `yaml:"regions,omitempty"`

	// TagCriteria defines tag validation rules specific to this resource type
	TagCriteria TagCriteria `yaml:"tag_criteria"`

	// ExcludedResources lists specific resources to be excluded from tag inspection
	ExcludedResources []ExcludedResource `yaml:"excluded_resources"`
}

// ExcludedResource defines a specific resource to be excluded from tag inspection,
// with a pattern to match and a reason for exclusion.
type ExcludedResource struct {
	// Pattern is a regex or identifier to match resources for exclusion
	Pattern string `yaml:"pattern"`

	// Reason explains why the resource is being excluded from tag inspection
	Reason string `yaml:"reason"`
}

// ComplianceLevel specifies the tag requirements for achieving a particular
// compliance status or level within the tag inspection process.
type ComplianceLevel struct {
	// RequiredTags is a list of tag keys that must be present to meet this compliance level
	RequiredTags []string `yaml:"required_tags"`

	// SpecificTags defines exact tag key-value pairs required for this compliance level
	SpecificTags map[string]string `yaml:"specific_tags"`
}

// CaseType represents the type of case validation
type CaseType string

const (
	CaseLowercase CaseType = "lowercase"
	CaseUppercase CaseType = "uppercase"
	CaseMixed     CaseType = "mixed"
)

// CaseRule defines the case validation rule for a tag
type CaseRule struct {
	Case    CaseType `yaml:"case"`
	Pattern string   `yaml:"pattern,omitempty"` // Optional pattern for mixed case
	Message string   `yaml:"message"`
}

// TagValidation contains all tag validation rules
type TagValidation struct {
	AllowedValues map[string][]string       `yaml:"allowed_values"`
	PatternRules  map[string]string         `yaml:"pattern_rules"`
	CaseRules     map[string]CaseRule       `yaml:"case_rules"`
	compiledRules map[string]*regexp.Regexp // Internal use for compiled patterns
}

// NotificationConfig manages the notification settings for reporting
// tag inspection results through different channels.
type NotificationConfig struct {
	// Slack contains configuration for Slack notifications
	Slack SlackNotificationConfig `yaml:"slack"`

	// Email contains configuration for email notifications
	Email EmailNotificationConfig `yaml:"email"`

	// Frequency determines how often notifications are sent
	Frequency string `yaml:"frequency"`
}

// SlackNotificationConfig defines the configuration for Slack notifications,
// including whether they are enabled and which channels to use.
type SlackNotificationConfig struct {
	// Enabled determines if Slack notifications are active
	Enabled bool `yaml:"enabled"`

	// Channels maps notification types to specific Slack channels
	Channels map[string]string `yaml:"channels"`
}

// EmailNotificationConfig specifies the email notification settings,
// including whether email notifications are enabled and the list of recipients.
type EmailNotificationConfig struct {
	// Enabled determines if email notifications are active
	Enabled bool `yaml:"enabled"`

	// Recipients is a list of email addresses to receive notifications
	Recipients []string `yaml:"recipients"`

	// Frequency determines how often email notifications are sent
	Frequency string `yaml:"frequency"`
}

// TagCriteria defines the criteria for validating resource tags in AWS.
// It allows specifying required, forbidden, and specific tag requirements.
type TagCriteria struct {
	// MinimumRequiredTags specifies the minimum number of tags that must be present
	MinimumRequiredTags int `yaml:"minimum_required_tags"`

	// RequiredTags is a list of tag keys that must be present on the resource
	RequiredTags []string `yaml:"required_tags"`

	// ForbiddenTags is a list of tag keys that must not be present on the resource
	ForbiddenTags []string `yaml:"forbidden_tags"`

	// SpecificTags is a map of tag key-value pairs that must exactly match
	SpecificTags map[string]string `yaml:"specific_tags"`

	// ComplianceLevel specifies the required compliance level for the resource
	ComplianceLevel string `yaml:"compliance_level"`
}

// Update the ComplianceLevel type or validation if needed
// For example, you might want to add a validation method
func IsValidComplianceLevel(level string) bool {
	validLevels := []string{"high", "medium", "low", "standard"}
	for _, validLevel := range validLevels {
		if level == validLevel {
			return true
		}
	}
	return false
}

// DefaultAWSRegion defines the default AWS region to use when no region is specified
const DefaultAWSRegion = "us-east-1"

// AWSConfig defines the AWS-specific configuration for region scanning
type AWSConfig struct {
	// Regions configuration for scanning
	Regions RegionsConfig `yaml:"regions"`

	// BatchSize specifies the number of resources to process in a single batch
	// If not set, it will fall back to the global batch size or a system default
	BatchSize *int `yaml:"batch_size,omitempty"`
}

// RegionsConfig specifies how AWS regions should be scanned
type RegionsConfig struct {
	// Mode determines the region scanning strategy
	// Can be 'all' to scan all regions or 'specific' to scan only listed regions
	Mode string `yaml:"mode"`

	// List of specific regions to scan when Mode is 'specific'
	List []string `yaml:"list,omitempty"`
}

// NormalizeAWSConfig ensures that AWS configuration has a valid configuration
func NormalizeAWSConfig(cfg *AWSConfig, globalCfg *GlobalConfig) {
	// If no AWS batch size is specified, use global batch size
	if cfg.BatchSize == nil {
		if globalCfg != nil && globalCfg.BatchSize != nil {
			cfg.BatchSize = globalCfg.BatchSize
		} else {
			// Default batch size if neither AWS nor global is set
			defaultBatchSize := 20
			cfg.BatchSize = &defaultBatchSize
		}
	}

	// If no AWS regions configuration is set, default to us-east-1
	if cfg.Regions.Mode == "" {
		cfg.Regions.Mode = "specific"
		cfg.Regions.List = []string{DefaultAWSRegion}
	}
}

// ValidAWSRegions provides a comprehensive list of valid AWS regions
func ValidAWSRegions() []string {
	return []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"ca-central-1",
		"eu-central-1", "eu-west-1", "eu-west-2", "eu-west-3", "eu-north-1",
		"ap-northeast-1", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2",
		"ap-south-1",
		"sa-east-1",
		"me-south-1",
		"af-south-1",
	}
}

// IsValidRegion checks if a given region is valid
func IsValidRegion(region string) bool {
	validRegions := ValidAWSRegions()
	for _, validRegion := range validRegions {
		if region == validRegion {
			return true
		}
	}
	return false
}
