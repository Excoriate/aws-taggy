package configuration

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Excoriate/aws-taggy/pkg/util"
	"github.com/xeipuuv/gojsonschema"
)

// FileValidator is responsible for validating configuration file paths and their existence.
type FileValidator struct {
	cfgPath string
}

// ContentValidator handles the validation of configuration content.
type ContentValidator struct {
	cfg *TaggyScanConfig
}

// NewFileValidator creates a new instance of FileValidator.
func NewFileValidator(cfgPath string) (*FileValidator, error) {
	if cfgPath == "" {
		return nil, fmt.Errorf("configuration file path cannot be empty")
	}
	return &FileValidator{cfgPath: cfgPath}, nil
}

// NewContentValidator creates a new instance of ContentValidator.
func NewContentValidator(cfg *TaggyScanConfig) (*ContentValidator, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}
	return &ContentValidator{cfg: cfg}, nil
}

// Validate performs comprehensive file validation
func (v *FileValidator) Validate() error {
	absPath, err := util.ResolveAbsolutePath(v.cfgPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	if err := util.FileExists(absPath); err != nil {
		return fmt.Errorf("configuration file does not exist: %w", err)
	}

	if err := util.FileHasExtension(absPath, ".yaml"); err != nil {
		return fmt.Errorf("configuration file has invalid extension: %w", err)
	}

	if err := util.FileIsNotEmpty(absPath); err != nil {
		return fmt.Errorf("configuration file is empty: %w", err)
	}

	return nil
}

// ValidateContent performs comprehensive validation of the configuration content
func (v *ContentValidator) ValidateContent() error {
	if err := v.validateAgainstSchema(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if err := v.validateVersion(); err != nil {
		return fmt.Errorf("version validation failed: %w", err)
	}

	if err := v.validateAWSConfig(); err != nil {
		return fmt.Errorf("AWS configuration validation failed: %w", err)
	}

	if err := v.validateGlobalConfig(); err != nil {
		return fmt.Errorf("global configuration validation failed: %w", err)
	}

	if err := v.validateResourceConfigs(); err != nil {
		return fmt.Errorf("resource configuration validation failed: %w", err)
	}

	if err := v.validateComplianceLevels(); err != nil {
		return fmt.Errorf("compliance levels validation failed: %w", err)
	}

	if err := v.validateTagValidation(); err != nil {
		return fmt.Errorf("tag validation configuration failed: %w", err)
	}

	if err := v.validateNotifications(); err != nil {
		return fmt.Errorf("notifications configuration failed: %w", err)
	}

	return nil
}

func (v *ContentValidator) validateAgainstSchema() error {
	// Load schema
	schemaLoader := gojsonschema.NewReferenceLoader("file://pkg/configuration/schema/tag-compliance-schema.json")

	// Convert config to JSON for validation
	configJSON, err := json.Marshal(v.cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// If version is empty, add a default version
	if v.cfg.Version == "" {
		v.cfg.Version = "1.0"
	}

	documentLoader := gojsonschema.NewBytesLoader(configJSON)

	// Perform validation
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, err := range result.Errors() {
			errors = append(errors, err.String())
		}
		return fmt.Errorf("configuration does not match schema: %s", strings.Join(errors, "; "))
	}

	return nil
}

func (v *ContentValidator) validateVersion() error {
	// If version is empty, set a default version
	if v.cfg.Version == "" {
		v.cfg.Version = "1.0"
		return nil
	}

	// Remove any quotes and whitespace from the version string
	version := strings.Trim(strings.TrimSpace(v.cfg.Version), `"'`)

	versionPattern := regexp.MustCompile(`^\d+\.\d+$`)
	if !versionPattern.MatchString(version) {
		return fmt.Errorf("invalid version format: %s, expected format: X.Y", version)
	}

	return nil
}

func (v *ContentValidator) validateAWSConfig() error {
	if v.cfg.AWS.Regions.Mode == "" {
		return fmt.Errorf("AWS regions mode is required")
	}

	if v.cfg.AWS.Regions.Mode != "all" && v.cfg.AWS.Regions.Mode != "specific" {
		return fmt.Errorf("invalid AWS regions mode: %s, expected: all or specific", v.cfg.AWS.Regions.Mode)
	}

	if v.cfg.AWS.Regions.Mode == "specific" && len(v.cfg.AWS.Regions.List) == 0 {
		return fmt.Errorf("specific AWS regions mode requires at least one region")
	}

	if v.cfg.AWS.BatchSize != nil && *v.cfg.AWS.BatchSize < 1 {
		return fmt.Errorf("AWS batch size must be greater than 0")
	}

	return nil
}

func (v *ContentValidator) validateGlobalConfig() error {
	if v.cfg.Global.BatchSize != nil && *v.cfg.Global.BatchSize <= 0 {
		return fmt.Errorf("global batch size must be positive")
	}

	if err := v.validateTagCriteria(v.cfg.Global.TagCriteria, "global"); err != nil {
		return err
	}

	return nil
}

func (v *ContentValidator) validateTagCriteria(criteria TagCriteria, context string) error {
	if criteria.MinimumRequiredTags < 0 {
		return fmt.Errorf("%s minimum required tags cannot be negative", context)
	}

	if len(criteria.RequiredTags) > 0 && criteria.MinimumRequiredTags > len(criteria.RequiredTags) {
		return fmt.Errorf("%s minimum required tags (%d) cannot exceed number of required tags (%d)",
			context, criteria.MinimumRequiredTags, len(criteria.RequiredTags))
	}

	if criteria.ComplianceLevel != "" && !v.isValidComplianceLevel(criteria.ComplianceLevel) {
		return fmt.Errorf("%s invalid compliance level: %s", context, criteria.ComplianceLevel)
	}

	return nil
}

// validateResourceType checks if the resource type is a supported AWS resource
func (v *ContentValidator) validateResourceType(resourceType string) error {
	supportedResources := map[string]bool{
		"ec2":             true,
		"s3":              true,
		"rds":             true,
		"lambda":          true,
		"dynamodb":        true,
		"elasticache":     true,
		"elb":             true,
		"alb":             true,
		"nlb":             true,
		"ebs":             true,
		"efs":             true,
		"eks":             true,
		"ecr":             true,
		"cloudfront":      true,
		"apigateway":      true,
		"route53":         true,
		"cloudwatch":      true,
		"sns":             true,
		"sqs":             true,
		"vpc":             true,
		"subnet":          true,
		"securitygroup":   true,
		"nacl":            true,
		"internetgateway": true,
		"natgateway":      true,
	}

	if !supportedResources[resourceType] {
		return fmt.Errorf("unsupported AWS resource type: %s", resourceType)
	}

	return nil
}

// validateResourceConfigs performs validation of resource configurations
func (v *ContentValidator) validateResourceConfigs() error {
	for resourceType, config := range v.cfg.Resources {
		if err := v.validateResourceType(resourceType); err != nil {
			return err
		}

		if !config.Enabled {
			continue
		}

		if err := v.validateTagCriteria(config.TagCriteria, fmt.Sprintf("resource %s", resourceType)); err != nil {
			return err
		}

		// Validate resource-specific compliance level against defined levels
		if config.TagCriteria.ComplianceLevel != "" {
			if _, exists := v.cfg.ComplianceLevels[config.TagCriteria.ComplianceLevel]; !exists {
				return fmt.Errorf("resource %s references undefined compliance level: %s",
					resourceType, config.TagCriteria.ComplianceLevel)
			}
		}

		for _, excluded := range config.ExcludedResources {
			if excluded.Pattern == "" {
				return fmt.Errorf("resource %s has empty exclusion pattern", resourceType)
			}
			if _, err := regexp.Compile(excluded.Pattern); err != nil {
				return fmt.Errorf("resource %s has invalid exclusion pattern: %s", resourceType, err)
			}
		}
	}

	return nil
}

func (v *ContentValidator) validateComplianceLevels() error {
	validLevels := map[string]bool{"high": true, "medium": true, "low": true, "standard": true}

	for level, config := range v.cfg.ComplianceLevels {
		if !validLevels[level] {
			return fmt.Errorf("invalid compliance level: %s", level)
		}

		if len(config.RequiredTags) == 0 && len(config.SpecificTags) == 0 {
			return fmt.Errorf("compliance level %s must define either required tags or specific tags", level)
		}
	}

	return nil
}

func (v *ContentValidator) validateTagValidation() error {
	if v.cfg.TagValidation.CaseRules != nil {
		for tag, rule := range v.cfg.TagValidation.CaseRules {
			if rule.Case == "" {
				return fmt.Errorf("case rule for tag %s must specify case type", tag)
			}
			if !v.isValidCaseType(rule.Case) {
				return fmt.Errorf("invalid case type for tag %s: %s", tag, rule.Case)
			}
			if rule.Pattern != "" {
				if _, err := regexp.Compile(rule.Pattern); err != nil {
					return fmt.Errorf("invalid pattern for tag %s: %s", tag, err)
				}
			}
		}
	}

	// Validate key validation rules
	if err := v.validateKeyValidation(); err != nil {
		return fmt.Errorf("key validation failed: %w", err)
	}

	// Validate value validation rules
	if err := v.validateValueValidation(); err != nil {
		return fmt.Errorf("value validation failed: %w", err)
	}

	if v.cfg.TagValidation.PatternRules != nil {
		for tag, pattern := range v.cfg.TagValidation.PatternRules {
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid pattern rule for tag %s: %s", tag, err)
			}
		}
	}

	if v.cfg.TagValidation.AllowedValues != nil {
		for tag, values := range v.cfg.TagValidation.AllowedValues {
			if len(values) == 0 {
				return fmt.Errorf("no allowed values specified for tag %s", tag)
			}
		}
	}

	// Validate length rules
	if err := v.validateLengthRules(); err != nil {
		return fmt.Errorf("length rules validation failed: %w", err)
	}

	return nil
}

func (v *ContentValidator) validateKeyValidation() error {
	keyValidation := v.cfg.TagValidation.KeyValidation

	// Validate max length
	if keyValidation.MaxLength <= 0 {
		return fmt.Errorf("key validation max length must be positive")
	}

	// Validate prefixes and suffixes
	for _, prefix := range keyValidation.AllowedPrefixes {
		if prefix == "" {
			return fmt.Errorf("empty prefix in allowed prefixes")
		}
	}

	for _, suffix := range keyValidation.AllowedSuffixes {
		if suffix == "" {
			return fmt.Errorf("empty suffix in allowed suffixes")
		}
	}

	return nil
}

func (v *ContentValidator) validateValueValidation() error {
	valueValidation := v.cfg.TagValidation.ValueValidation

	// Validate allowed characters pattern
	if valueValidation.AllowedCharacters != "" {
		if _, err := regexp.Compile(fmt.Sprintf("[%s]", valueValidation.AllowedCharacters)); err != nil {
			return fmt.Errorf("invalid allowed characters pattern: %s", err)
		}
	}

	// Validate disallowed values
	for _, value := range valueValidation.DisallowedValues {
		if value == "" {
			return fmt.Errorf("empty value in disallowed values")
		}
	}

	return nil
}

func (v *ContentValidator) validateLengthRules() error {
	for tag, rule := range v.cfg.TagValidation.LengthRules {
		if rule.MinLength != nil && *rule.MinLength < 0 {
			return fmt.Errorf("tag %s has negative minimum length", tag)
		}
		if rule.MaxLength != nil && rule.MinLength != nil && *rule.MaxLength <= *rule.MinLength {
			return fmt.Errorf("tag %s has maximum length (%d) less than or equal to minimum length (%d)",
				tag, *rule.MaxLength, *rule.MinLength)
		}
	}
	return nil
}

func (v *ContentValidator) validateNotifications() error {
	if v.cfg.Notifications.Slack.Enabled {
		if len(v.cfg.Notifications.Slack.Channels) == 0 {
			return fmt.Errorf("slack notifications enabled but no channels configured")
		}
	}

	if v.cfg.Notifications.Email.Enabled {
		if len(v.cfg.Notifications.Email.Recipients) == 0 {
			return fmt.Errorf("email notifications enabled but no recipients configured")
		}
		for _, email := range v.cfg.Notifications.Email.Recipients {
			if !v.isValidEmail(email) {
				return fmt.Errorf("invalid email address: %s", email)
			}
		}
		if v.cfg.Notifications.Email.Frequency == "" {
			return fmt.Errorf("email notifications enabled but no frequency specified")
		}
		if !v.isValidEmailFrequency(v.cfg.Notifications.Email.Frequency) {
			return fmt.Errorf("invalid email frequency: %s", v.cfg.Notifications.Email.Frequency)
		}
	}

	return nil
}

func (v *ContentValidator) isValidComplianceLevel(level string) bool {
	validLevels := map[string]bool{
		"high":     true,
		"medium":   true,
		"low":      true,
		"standard": true,
	}
	return validLevels[level]
}

func (v *ContentValidator) isValidCaseType(caseType CaseType) bool {
	return caseType == CaseLowercase ||
		caseType == CaseUppercase ||
		caseType == CaseMixed
}

func (v *ContentValidator) isValidEmail(email string) bool {
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailPattern.MatchString(email)
}

func (v *ContentValidator) isValidEmailFrequency(freq string) bool {
	validFrequencies := map[string]bool{
		"daily":  true,
		"hourly": true,
		"weekly": true,
	}
	return validFrequencies[freq]
}
