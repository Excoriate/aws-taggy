package configuration

import (
	"fmt"
)

// ConfigQuerier provides methods to retrieve specific parts of the configuration
type ConfigQuerier struct {
	config *TaggyScanConfig
}

// NewConfigQuerier creates a new ConfigQuerier with the provided configuration
func NewConfigQuerier(config *TaggyScanConfig) (*ConfigQuerier, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}
	return &ConfigQuerier{config: config}, nil
}

// GetResources retrieves all resource configurations
//
// Returns:
//   - A map of resource configurations
//   - An error if no resources are configured
func (q *ConfigQuerier) GetResources() (map[string]ResourceConfig, error) {
	if len(q.config.Resources) == 0 {
		return nil, fmt.Errorf("no resources configured")
	}
	return q.config.Resources, nil
}

// GetAWSConfig retrieves the AWS configuration
//
// Returns:
//   - The AWS configuration
//   - An error if AWS configuration is not set
func (q *ConfigQuerier) GetAWSConfig() (*AWSConfig, error) {
	if q.config.AWS.Regions.Mode == "" {
		return nil, fmt.Errorf("AWS configuration is not set")
	}
	return &q.config.AWS, nil
}

// GetComplianceLevels retrieves all compliance level configurations
//
// Returns:
//   - A map of compliance levels
//   - An error if no compliance levels are configured
func (q *ConfigQuerier) GetComplianceLevels() (map[string]ComplianceLevel, error) {
	if len(q.config.ComplianceLevels) == 0 {
		return nil, fmt.Errorf("no compliance levels configured")
	}
	return q.config.ComplianceLevels, nil
}

// GetTagValidationConfig retrieves the tag validation configuration
//
// Returns:
//   - The tag validation configuration
//   - An error if tag validation configuration is not set
func (q *ConfigQuerier) GetTagValidationConfig() (*TagValidation, error) {
	if len(q.config.TagValidation.AllowedValues) == 0 &&
		len(q.config.TagValidation.PatternRules) == 0 {
		return nil, fmt.Errorf("tag validation configuration is not set")
	}
	return &q.config.TagValidation, nil
}

// GetNotificationsConfig retrieves the notifications configuration
//
// Returns:
//   - The notifications configuration
//   - An error if notifications configuration is not set
func (q *ConfigQuerier) GetNotificationsConfig() (*NotificationConfig, error) {
	// Check if either Slack or Email notifications are enabled
	if !q.config.Notifications.Slack.Enabled &&
		!q.config.Notifications.Email.Enabled {
		return nil, fmt.Errorf("notifications configuration is not set")
	}
	return &q.config.Notifications, nil
}

// GetResourceByType retrieves a specific resource configuration by its type
//
// Parameters:
//   - resourceType: The type of resource to retrieve (e.g., "s3", "ec2")
//
// Returns:
//   - The resource configuration for the specified type
//   - An error if the resource type is not found
func (q *ConfigQuerier) GetResourceByType(resourceType string) (*ResourceConfig, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resource type cannot be empty")
	}

	resourceConfig, exists := q.config.Resources[resourceType]
	if !exists {
		return nil, fmt.Errorf("resource type %s not found in configuration", resourceType)
	}

	return &resourceConfig, nil
}

// GetComplianceLevelByName retrieves a specific compliance level configuration
//
// Parameters:
//   - levelName: The name of the compliance level to retrieve
//
// Returns:
//   - The compliance level configuration
//   - An error if the compliance level is not found
func (q *ConfigQuerier) GetComplianceLevelByName(levelName string) (*ComplianceLevel, error) {
	if levelName == "" {
		return nil, fmt.Errorf("compliance level name cannot be empty")
	}

	complianceLevel, exists := q.config.ComplianceLevels[levelName]
	if !exists {
		return nil, fmt.Errorf("compliance level %s not found in configuration", levelName)
	}

	return &complianceLevel, nil
}

// GetResourceRegions retrieves the regions for a specific resource type
//
// Parameters:
//   - resourceType: The type of resource to retrieve (e.g., "s3", "ec2")
//
// Returns:
//   - A list of regions for the specified resource type
//   - An error if the resource type is not found
func (q *ConfigQuerier) GetResourceRegions(resourceType string) ([]string, error) {
	resource, err := q.GetResourceByType(resourceType)
	if err != nil {
		return nil, err
	}

	// If resource-specific regions are set, return those
	if len(resource.Regions) > 0 {
		return resource.Regions, nil
	}

	// If no resource-specific regions, return AWS config regions
	awsConfig, err := q.GetAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS configuration: %w", err)
	}

	// If AWS regions mode is 'all', return all valid AWS regions
	if awsConfig.Regions.Mode == "all" {
		return ValidAWSRegions(), nil
	}

	// Return the specific regions from AWS config
	return awsConfig.Regions.List, nil
}
