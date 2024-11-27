package configuration

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Excoriate/aws-taggy/pkg/constants"
	"github.com/Excoriate/aws-taggy/pkg/util"
)

// ConfigFileValidator is responsible for validating configuration file paths and their existence.
// It provides methods to check the validity of a configuration file before processing.
type ConfigFileValidator struct {
	cfgPath string
}

// ConfigValidator handles the validation of configuration content.
// It ensures that the configuration meets the required criteria for further processing.
type ConfigValidator struct {
	cfg *TaggyScanConfig
}

// NewConfigFileValidator creates a new instance of ConfigFileValidator.
// It validates the input configuration file path and returns an error if the path is invalid.
//
// Parameters:
//   - cfgPath: The file path of the configuration file to be validated.
//
// Returns:
//   - A pointer to the created ConfigFileValidator
//   - An error if the configuration file path is empty
func NewConfigFileValidator(cfgPath string) (*ConfigFileValidator, error) {
	if cfgPath == "" {
		return nil, fmt.Errorf("configuration file path cannot be empty")
	}

	return &ConfigFileValidator{cfgPath: cfgPath}, nil
}

// NewConfigValidator creates a new instance of ConfigValidator.
// It validates the input configuration and ensures it is not nil.
//
// Parameters:
//   - cfg: The configuration object to be validated.
//
// Returns:
//   - A pointer to the created ConfigValidator
//   - An error if the configuration is nil
func NewConfigValidator(cfg *TaggyScanConfig) (*ConfigValidator, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	return &ConfigValidator{cfg: cfg}, nil
}

// Validate performs comprehensive file validation
// Validate performs a comprehensive validation of the configuration file.
// It performs multiple checks to ensure the configuration file is valid and ready for processing.
//
// The validation steps include:
//  1. Resolving the absolute path of the configuration file
//  2. Checking if the configuration file exists
//  3. Validating the file extension (expecting .yaml)
//  4. Ensuring the configuration file is not empty
//
// Returns:
//   - An error if any validation step fails, otherwise nil
func (v *ConfigFileValidator) Validate() error {
	// Resolve absolute path to ensure the file path is fully qualified
	_, err := util.ResolveAbsolutePath(v.cfgPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check file existence to confirm the configuration file is present
	if err := v.ValidateConfigFileExist(); err != nil {
		return err
	}

	// Validate file extension to ensure it's a supported configuration file type
	if err := v.ValidateConfigFileHasExtension(); err != nil {
		return err
	}

	// Verify the configuration file is not empty to prevent processing blank files
	if err := v.ValidateConfigFileIsNotEmpty(); err != nil {
		return err
	}

	return nil
}

// ValidateConfigFileExist checks if the configuration file exists on the filesystem.
//
// This method verifies the presence of the configuration file specified by the validator's
// configuration path. It uses the FileExists utility function to perform the check.
//
// Returns:
//   - nil if the file exists successfully
//   - An error if the file cannot be found, wrapping the underlying filesystem error
func (v *ConfigFileValidator) ValidateConfigFileExist() error {
	if err := util.FileExists(v.cfgPath); err != nil {
		return fmt.Errorf("configuration file does not exist: %w", err)
	}

	return nil
}

// ValidateConfigFileIsNotEmpty checks if the configuration file is not empty.
//
// This method verifies that the configuration file specified by the validator's
// configuration path is not empty. It uses the FileIsNotEmpty utility function to perform the check.
//
// Returns:
//   - nil if the file is not empty successfully
//   - An error if the file is empty, wrapping the underlying filesystem error
func (v *ConfigFileValidator) ValidateConfigFileIsNotEmpty() error {
	if err := util.FileIsNotEmpty(v.cfgPath); err != nil {
		return fmt.Errorf("configuration file is empty: %w", err)
	}

	return nil
}

// ValidateConfigFileHasExtension checks if the configuration file has the correct extension.
//
// This method verifies that the configuration file specified by the validator's
// configuration path has the correct extension (".yaml"). It uses the FileHasExtension utility function to perform the check.
//
// Returns:
//   - nil if the file has the correct extension successfully
//   - An error if the file does not have the correct extension, wrapping the underlying error
func (v *ConfigFileValidator) ValidateConfigFileHasExtension() error {
	if err := util.FileHasExtension(v.cfgPath, ".yaml"); err != nil {
		return fmt.Errorf("configuration file has invalid extension: %w", err)
	}

	return nil
}

// ValidateVersion checks the configuration version
// ValidateVersion checks and validates the configuration version.
//
// This method performs two primary validation checks on the configuration version:
//  1. Ensures that the version is not an empty string
//  2. Verifies that the version matches the currently supported configuration version
//
// The method checks against a predefined constant (SupportedConfigVersion) to ensure
// compatibility with the expected configuration format.
//
// Returns:
//   - nil if the version is valid and supported
//   - An error if:
//   - The version is an empty string
//   - The version does not match the supported version
//
// Example error scenarios:
//   - "configuration version is missing"
//   - "unsupported configuration version: 0.1.0. Expected 1.0.0"
func (c *ConfigValidator) ValidateVersion() error {
	if c.cfg.Version == "" {
		return fmt.Errorf("configuration version is missing")
	}

	// Strict version check
	if c.cfg.Version != constants.SupportedConfigVersion {
		return fmt.Errorf("unsupported configuration version: %s. Expected %s",
			c.cfg.Version, constants.SupportedConfigVersion)
	}

	return nil
}

// ValidateGlobalConfig validates global configuration settings
// ValidateGlobalConfig validates the global configuration settings for the scanner.
//
// This method performs comprehensive validation on global configuration parameters,
// ensuring that the global settings meet the required constraints before processing.
// It checks two primary aspects of the global configuration:
//  1. Batch Size: Ensures that if a batch size is specified, it is a positive number
//  2. Tag Criteria: Validates the tag criteria using a separate validation method
//
// The method performs the following specific validations:
//   - Checks that the global batch size (if set) is a positive integer
//   - Validates the tag criteria using the validateTagCriteria method
//
// Returns:
//   - nil if all global configuration settings are valid
//   - An error with a descriptive message if any validation fails, including:
//   - Invalid batch size (non-positive number)
//   - Invalid tag criteria
//
// Example error scenarios:
//   - "global batch size must be a positive number"
//   - "global tag criteria validation failed: ..."
func (v *ConfigValidator) ValidateGlobalConfig() error {
	// Validate batch size
	if v.cfg.Global.BatchSize != nil && *v.cfg.Global.BatchSize <= 0 {
		return fmt.Errorf("global batch size must be a positive number")
	}

	// Validate tag criteria
	if err := v.validateTagCriteria(v.cfg.Global.TagCriteria); err != nil {
		return fmt.Errorf("global tag criteria validation failed: %w", err)
	}

	return nil
}

// ValidateResourceConfigs validates resource-specific configurations
func (v *ConfigValidator) ValidateResourceConfigs() error {
	// Validate each resource configuration
	for resourceType, resourceConfig := range v.cfg.Resources {
		// Validate that the resource is enabled before further checks
		if !resourceConfig.Enabled {
			continue
		}

		// Validate tag criteria for the resource
		if err := v.validateTagCriteria(resourceConfig.TagCriteria); err != nil {
			return fmt.Errorf("invalid tag criteria for resource type %s: %w", resourceType, err)
		}

		// Validate excluded resources
		for _, excludedResource := range resourceConfig.ExcludedResources {
			if excludedResource.Pattern == "" {
				return fmt.Errorf("excluded resource pattern cannot be empty for resource type %s", resourceType)
			}
		}

		// Validate compliance level
		if resourceConfig.TagCriteria.ComplianceLevel != "" {
			if !IsValidComplianceLevel(resourceConfig.TagCriteria.ComplianceLevel) {
				return fmt.Errorf("invalid compliance level %s for resource type %s",
					resourceConfig.TagCriteria.ComplianceLevel, resourceType)
			}
		}
	}

	return nil
}

// ValidateComplianceLevels validates compliance level configurations
// ValidateComplianceLevels performs comprehensive validation of compliance level configurations.
//
// This method validates the integrity and completeness of compliance levels defined in the configuration.
// It ensures that:
//  1. Each compliance level has a non-empty name
//  2. Required tags within each compliance level are not empty
//  3. Specific tags within each compliance level have non-empty keys and values
//
// The validation process checks the following constraints:
//   - Compliance level names must be non-empty strings
//   - Required tags cannot be blank
//   - Specific tags must have both non-empty keys and values
//
// Returns:
//   - nil if all compliance levels pass validation
//   - An error with a descriptive message if any validation fails, specifying:
//   - Empty compliance level name
//   - Empty required tag
//   - Empty key or value in specific tags
//
// Example error scenarios:
//   - "compliance level name cannot be empty"
//   - "empty required tag in compliance level 'production'"
//   - "empty key or value in specific tags of compliance level 'staging'"
func (v *ConfigValidator) ValidateComplianceLevels() error {
	validLevels := []string{"high", "medium", "low", "standard"}

	for levelName, level := range v.cfg.ComplianceLevels {
		// Check if the level name is valid
		found := false
		for _, validLevel := range validLevels {
			if levelName == validLevel {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid compliance level: %s. Must be one of %v",
				levelName, validLevels)
		}

		// Validate required tags
		for _, tag := range level.RequiredTags {
			if tag == "" {
				return fmt.Errorf("empty required tag in compliance level '%s'", levelName)
			}
		}

		// Validate specific tags
		for key, value := range level.SpecificTags {
			if key == "" || value == "" {
				return fmt.Errorf("empty key or value in specific tags of compliance level '%s'", levelName)
			}
		}
	}
	return nil
}

// ValidateTagValidationRules validates tag validation configuration
// ValidateTagValidationRules performs comprehensive validation of tag validation configuration.
//
// This method validates the integrity and completeness of tag validation rules, ensuring:
//  1. Allowed values for each tag are non-empty
//  2. Each allowed value is a non-empty string
//  3. Pattern rules for each tag are non-empty
//
// The validation process checks the following constraints:
//   - Tags must have at least one allowed value
//   - Allowed values cannot be blank strings
//   - Pattern rules must be non-empty strings
//
// Returns:
//   - nil if all tag validation rules pass validation
//   - An error with a descriptive message if any validation fails, specifying:
//   - Tags with no allowed values
//   - Empty values in allowed values list
//   - Empty pattern rules
//
// Example error scenarios:
//   - "no allowed values specified for tag environment"
//   - "empty value found in allowed values for tag team"
//   - "empty pattern rule for tag service"
func (v *ConfigValidator) ValidateTagValidationRules() error {
	// Validate allowed values
	for tagName, values := range v.cfg.TagValidation.AllowedValues {
		if len(values) == 0 {
			return fmt.Errorf("no allowed values specified for tag %s", tagName)
		}

		// Validate each value is not empty
		for _, value := range values {
			if value == "" {
				return fmt.Errorf("empty value found in allowed values for tag %s", tagName)
			}
		}
	}

	// Validate pattern rules
	for tagName, pattern := range v.cfg.TagValidation.PatternRules {
		if pattern == "" {
			return fmt.Errorf("empty pattern rule for tag %s", tagName)
		}
	}

	return nil
}

// ValidateNotifications validates notification configurations
// ValidateNotifications performs comprehensive validation of notification configurations.
//
// This method validates the integrity and completeness of notification settings, ensuring:
//  1. Slack notifications have at least one channel when enabled
//  2. Email notifications have at least one recipient when enabled
//  3. Email notification frequency is valid if specified
//
// The validation process checks the following constraints:
//   - Slack notifications require at least one channel when enabled
//   - Email notifications require at least one recipient when enabled
//   - Email notification frequency must be one of: "daily", "hourly", or "weekly"
//
// Returns:
//   - nil if all notification configurations pass validation
//   - An error with a descriptive message if any validation fails, specifying:
//   - Missing Slack channels when Slack notifications are enabled
//   - Missing email recipients when email notifications are enabled
//   - Invalid email notification frequency
//
// Example error scenarios:
//   - "Slack notifications enabled but no channels specified"
//   - "email notifications enabled but no recipients specified"
//   - "invalid email notification frequency: monthly"
func (v *ConfigValidator) ValidateNotifications() error {
	// Validate Slack notifications
	if v.cfg.Notifications.Slack.Enabled {
		if len(v.cfg.Notifications.Slack.Channels) == 0 {
			return fmt.Errorf("Slack notifications enabled but no channels specified")
		}
	}

	// Validate Email notifications
	if v.cfg.Notifications.Email.Enabled {
		if len(v.cfg.Notifications.Email.Recipients) == 0 {
			return fmt.Errorf("email notifications enabled but no recipients specified")
		}

		// Validate email frequency
		validFrequencies := []string{"daily", "hourly", "weekly"}
		if v.cfg.Notifications.Email.Frequency != "" {
			found := false
			for _, freq := range validFrequencies {
				if v.cfg.Notifications.Email.Frequency == freq {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid email notification frequency: %s", v.cfg.Notifications.Email.Frequency)
			}
		}
	}

	return nil
}

// ValidateAWSConfig validates AWS configuration
func (v *ConfigValidator) ValidateAWSConfig() error {
	// Ensure AWS configuration is normalized
	NormalizeAWSConfig(&v.cfg.AWS, &v.cfg.Global)

	// Validate regions mode
	if v.cfg.AWS.Regions.Mode != "all" && v.cfg.AWS.Regions.Mode != "specific" {
		return fmt.Errorf("invalid AWS regions mode. Must be 'all' or 'specific', got %s", v.cfg.AWS.Regions.Mode)
	}

	// Validate batch size
	if v.cfg.AWS.BatchSize != nil && *v.cfg.AWS.BatchSize <= 0 {
		return fmt.Errorf("AWS batch size must be a positive number")
	}

	// If mode is 'specific', validate the list of regions
	if v.cfg.AWS.Regions.Mode == "specific" {
		// Check if list is empty when mode is 'specific'
		if len(v.cfg.AWS.Regions.List) == 0 {
			// If empty, set to default region
			v.cfg.AWS.Regions.List = []string{DefaultAWSRegion}
		}

		// Validate each specified region
		for _, region := range v.cfg.AWS.Regions.List {
			if region == "" {
				return fmt.Errorf("empty region specified in AWS regions list")
			}

			if !IsValidRegion(region) {
				return fmt.Errorf("invalid AWS region: %s", region)
			}
		}
	}

	return nil
}

// Validate performs a comprehensive validation of the entire configuration
// This method validates the configuration that was passed during initialization
//
// Returns:
//   - An error if any validation fails, otherwise nil
func (v *ConfigValidator) Validate() error {
	// Validate version
	if err := v.ValidateVersion(); err != nil {
		return fmt.Errorf("version validation failed: %w", err)
	}

	// Validate global configuration
	if err := v.ValidateGlobalConfig(); err != nil {
		return fmt.Errorf("global configuration validation failed: %w", err)
	}

	// Validate resource configurations
	if err := v.ValidateResourceConfigs(); err != nil {
		return fmt.Errorf("resource configuration validation failed: %w", err)
	}

	// Validate compliance levels
	if err := v.ValidateComplianceLevels(); err != nil {
		return fmt.Errorf("compliance levels validation failed: %w", err)
	}

	// Validate tag validation rules
	if err := v.ValidateTagValidationRules(); err != nil {
		return fmt.Errorf("tag validation rules validation failed: %w", err)
	}

	// Validate notifications
	if err := v.ValidateNotifications(); err != nil {
		return fmt.Errorf("notifications validation failed: %w", err)
	}

	// Validate AWS configuration
	if err := v.ValidateAWSConfig(); err != nil {
		return fmt.Errorf("AWS configuration validation failed: %w", err)
	}

	// Validate resource-specific regions
	if err := v.validateResourceRegions(); err != nil {
		return fmt.Errorf("resource regions validation failed: %w", err)
	}

	return nil
}

// validateTagCriteria is a helper method to validate tag criteria
// validateTagCriteria validates the tag criteria configuration for a resource or global setting.
//
// This method performs two key validations on the provided TagCriteria:
// 1. Ensures that the minimum number of required tags is not negative
// 2. Verifies that the minimum required tags does not exceed the total number of defined required tags
//
// Parameters:
//   - criteria: The TagCriteria configuration to be validated
//
// Returns:
//   - An error if the tag criteria configuration is invalid
//   - nil if the tag criteria configuration passes all validation checks
//
// Validation rules:
//   - MinimumRequiredTags must be a non-negative integer
//   - MinimumRequiredTags cannot be greater than the length of RequiredTags slice
func (v *ConfigValidator) validateTagCriteria(criteria TagCriteria) error {
	// Validate minimum required tags
	if criteria.MinimumRequiredTags < 0 {
		return fmt.Errorf("minimum required tags cannot be negative")
	}

	if criteria.MinimumRequiredTags > len(criteria.RequiredTags) {
		return fmt.Errorf("minimum required tags (%d) cannot be greater than the number of required tags (%d)",
			criteria.MinimumRequiredTags, len(criteria.RequiredTags))
	}

	return nil
}

// validateResourceRegions is a helper method to validate resource-specific regions
// validateResourceRegions validates the resource-specific regions configuration for all resources.
//
// This method performs validation on the resource-specific regions configuration for all resources.
// It ensures that if regions are specified for a resource, they are valid AWS regions.
//
// Returns:
//   - An error if any resource-specific region configuration is invalid
//   - nil if all resource-specific regions pass validation
func (v *ConfigValidator) validateResourceRegions() error {
	// Validate that if regions are specified for a resource, they are valid AWS regions
	for resourceType, resourceConfig := range v.cfg.Resources {
		if len(resourceConfig.Regions) > 0 {
			for _, region := range resourceConfig.Regions {
				if !IsValidRegion(region) {
					return fmt.Errorf("invalid region %s specified for resource type %s", region, resourceType)
				}
			}
		}
	}
	return nil
}

// ValidateTagCase validates the case of a tag value according to the case rules
func (v *ConfigValidator) ValidateTagCase(tagName, tagValue string) error {
	if v.cfg.TagValidation.CaseRules == nil {
		return nil
	}

	rule, exists := v.cfg.TagValidation.CaseRules[tagName]
	if !exists {
		return nil
	}

	switch rule.Case {
	case CaseLowercase:
		if tagValue != strings.ToLower(tagValue) {
			return fmt.Errorf(rule.Message)
		}
	case CaseUppercase:
		if tagValue != strings.ToUpper(tagValue) {
			return fmt.Errorf(rule.Message)
		}
	case CaseMixed:
		if rule.Pattern != "" {
			matched, err := regexp.MatchString(rule.Pattern, tagValue)
			if err != nil {
				return fmt.Errorf("invalid mixed case pattern for tag %s: %w", tagName, err)
			}
			if !matched {
				return fmt.Errorf(rule.Message)
			}
		}
	}

	return nil
}

// ValidateTag validates a single tag against all validation rules
func (v *ConfigValidator) ValidateTag(tagName, tagValue string) error {
	// Check case rules first
	if err := v.ValidateTagCase(tagName, tagValue); err != nil {
		return fmt.Errorf("case validation failed for tag %s: %w", tagName, err)
	}

	// Check allowed values
	if allowedValues, exists := v.cfg.TagValidation.AllowedValues[tagName]; exists {
		found := false
		for _, allowed := range allowedValues {
			if tagValue == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid value for tag %s: %s not in allowed values", tagName, tagValue)
		}
	}

	// Check pattern rules
	if pattern, exists := v.cfg.TagValidation.PatternRules[tagName]; exists {
		if v.cfg.TagValidation.compiledRules == nil {
			v.cfg.TagValidation.compiledRules = make(map[string]*regexp.Regexp)
		}

		regex, exists := v.cfg.TagValidation.compiledRules[tagName]
		if !exists {
			var err error
			regex, err = regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern for tag %s: %w", tagName, err)
			}
			v.cfg.TagValidation.compiledRules[tagName] = regex
		}

		if !regex.MatchString(tagValue) {
			return fmt.Errorf("tag %s value does not match required pattern: %s", tagName, pattern)
		}
	}

	return nil
}

// ValidateTagValidation performs comprehensive validation of tag validation rules
func (v *ConfigValidator) ValidateTagValidation() error {
	// Validate allowed values
	if err := v.validateAllowedValues(); err != nil {
		return err
	}

	// Validate case sensitivity configuration
	if err := v.validateCaseSensitivity(); err != nil {
		return err
	}

	// Validate case transformation rules
	if err := v.validateCaseRules(); err != nil {
		return err
	}

	return nil
}

// validateAllowedValues checks that allowed values are valid and non-empty
func (v *ConfigValidator) validateAllowedValues() error {
	for tagName, values := range v.cfg.TagValidation.AllowedValues {
		if len(values) == 0 {
			return fmt.Errorf("no allowed values specified for tag %s", tagName)
		}

		// Ensure no duplicate values
		uniqueValues := make(map[string]bool)
		for _, value := range values {
			if value == "" {
				return fmt.Errorf("empty value not allowed in allowed values for tag %s", tagName)
			}
			if uniqueValues[value] {
				return fmt.Errorf("duplicate value %s found for tag %s", value, tagName)
			}
			uniqueValues[value] = true
		}
	}
	return nil
}

// validateCaseSensitivity checks the case sensitivity configuration
func (v *ConfigValidator) validateCaseSensitivity() error {
	for tagName, caseSensitivity := range v.cfg.TagValidation.CaseSensitivity {
		// Validate case sensitivity mode
		switch caseSensitivity.Mode {
		case CaseValidationStrict, CaseValidationRelaxed:
			// Valid modes
		default:
			return fmt.Errorf("invalid case validation mode %s for tag %s", caseSensitivity.Mode, tagName)
		}

		// Ensure the tag has allowed values defined if using strict mode
		if caseSensitivity.Mode == CaseValidationStrict {
			if _, exists := v.cfg.TagValidation.AllowedValues[tagName]; !exists {
				return fmt.Errorf("strict case validation requires allowed values for tag %s", tagName)
			}
		}
	}
	return nil
}

// validateCaseRules checks the case transformation rules
func (v *ConfigValidator) validateCaseRules() error {
	for tagName, caseRule := range v.cfg.TagValidation.CaseRules {
		// Validate case type
		switch caseRule.Case {
		case CaseLowercase, CaseUppercase, CaseMixed:
			// Valid case types
		default:
			return fmt.Errorf("invalid case transformation %s for tag %s", caseRule.Case, tagName)
		}

		// For mixed case, validate pattern if provided
		if caseRule.Case == CaseMixed && caseRule.Pattern != "" {
			_, err := regexp.Compile(caseRule.Pattern)
			if err != nil {
				return fmt.Errorf("invalid mixed case pattern for tag %s: %w", tagName, err)
			}
		}

		// Ensure message is provided
		if caseRule.Message == "" {
			return fmt.Errorf("case rule for tag %s must have a message", tagName)
		}
	}
	return nil
}

// ValidateTagKey validates a tag key against all key-related rules
func (v *ConfigValidator) ValidateTagKey(key string) error {
	// Check key format rules
	for _, rule := range v.cfg.TagValidation.KeyFormatRules {
		matched, err := regexp.MatchString(rule.Pattern, key)
		if err != nil {
			return fmt.Errorf("invalid key format pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf(rule.Message)
		}
	}

	// Check prohibited tags
	for _, prohibited := range v.cfg.TagValidation.ProhibitedTags {
		if strings.HasPrefix(key, prohibited) {
			return fmt.Errorf("tag key '%s' uses prohibited prefix '%s'", key, prohibited)
		}
	}

	// Check key validation rules
	if v.cfg.TagValidation.KeyValidation.MaxLength > 0 && len(key) > v.cfg.TagValidation.KeyValidation.MaxLength {
		return fmt.Errorf("tag key exceeds maximum length of %d characters", v.cfg.TagValidation.KeyValidation.MaxLength)
	}

	// Check allowed prefixes
	if len(v.cfg.TagValidation.KeyValidation.AllowedPrefixes) > 0 {
		validPrefix := false
		for _, prefix := range v.cfg.TagValidation.KeyValidation.AllowedPrefixes {
			if strings.HasPrefix(key, prefix) {
				validPrefix = true
				break
			}
		}
		if !validPrefix {
			return fmt.Errorf("tag key must start with one of the allowed prefixes: %v", v.cfg.TagValidation.KeyValidation.AllowedPrefixes)
		}
	}

	return nil
}

// ValidateTagValue validates a tag value against all value-related rules
func (v *ConfigValidator) ValidateTagValue(key, value string) error {
	// Check length rules
	if rule, exists := v.cfg.TagValidation.LengthRules[key]; exists {
		if rule.MinLength != nil && len(value) < *rule.MinLength {
			return fmt.Errorf(rule.Message)
		}
		if rule.MaxLength != nil && len(value) > *rule.MaxLength {
			return fmt.Errorf(rule.Message)
		}
	}

	// Check allowed values
	if allowedValues, exists := v.cfg.TagValidation.AllowedValues[key]; exists {
		valid := false
		for _, allowed := range allowedValues {
			if value == allowed {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("value '%s' not in allowed values for tag '%s'", value, key)
		}
	}

	// Check pattern rules
	if pattern, exists := v.cfg.TagValidation.PatternRules[key]; exists {
		matched, err := regexp.MatchString(pattern, value)
		if err != nil {
			return fmt.Errorf("invalid pattern for tag %s: %w", key, err)
		}
		if !matched {
			return fmt.Errorf("value '%s' does not match required pattern for tag '%s'", value, key)
		}
	}

	// Check disallowed values
	for _, disallowed := range v.cfg.TagValidation.ValueValidation.DisallowedValues {
		if value == disallowed {
			return fmt.Errorf("value '%s' is not allowed for tag '%s'", value, key)
		}
	}

	// Check allowed characters
	if pattern := v.cfg.TagValidation.ValueValidation.AllowedCharacters; pattern != "" {
		matched, err := regexp.MatchString(fmt.Sprintf("^[%s]*$", pattern), value)
		if err != nil {
			return fmt.Errorf("invalid allowed characters pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("value contains disallowed characters")
		}
	}

	return nil
}

// ValidateTagCount validates the total number of tags
func (v *ConfigValidator) ValidateTagCount(tags map[string]string) error {
	if v.cfg.Global.TagCriteria.MaxTags > 0 && len(tags) > v.cfg.Global.TagCriteria.MaxTags {
		return fmt.Errorf("number of tags (%d) exceeds maximum allowed (%d)", len(tags), v.cfg.Global.TagCriteria.MaxTags)
	}
	return nil
}
