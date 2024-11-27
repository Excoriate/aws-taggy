package compliance

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
)

// Validator defines the interface for tag compliance validation
type Validator interface {
	ValidateTags(tags map[string]string) *ComplianceResult
}

// TagValidator implements the Validator interface
type TagValidator struct {
	config *configuration.TaggyScanConfig
}

// NewTagValidator creates a new TagValidator with the given configuration
func NewTagValidator(config *configuration.TaggyScanConfig) *TagValidator {
	return &TagValidator{
		config: config,
	}
}

// ValidateTags checks the compliance of a set of tags against the configuration
func (v *TagValidator) ValidateTags(tags map[string]string) *ComplianceResult {
	result := &ComplianceResult{
		IsCompliant:  true,
		Violations:   []Violation{},
		ResourceTags: tags,
	}

	// Check tag count limits
	if v.config.Global.TagCriteria.MaxTags > 0 {
		v.checkTagCount(tags, result)
	}

	// Check minimum required tags
	if v.config.Global.TagCriteria.MinimumRequiredTags > 0 {
		v.checkRequiredTags(tags, result)
	}

	// Check prohibited tags
	if len(v.config.TagValidation.ProhibitedTags) > 0 {
		v.checkProhibitedTags(tags, result)
	}

	// Validate tag key format
	if v.config.TagValidation.KeyFormatRules != nil {
		v.validateKeyFormat(tags, result)
	}

	// Validate case rules
	if v.config.TagValidation.CaseRules != nil {
		v.validateCaseRules(tags, result)
	}

	// Validate allowed values
	if v.config.TagValidation.AllowedValues != nil {
		v.validateAllowedValues(tags, result)
	}

	// Validate pattern rules
	if v.config.TagValidation.PatternRules != nil {
		v.validatePatternRules(tags, result)
	}

	// Validate value length
	if v.config.TagValidation.LengthRules != nil {
		v.validateValueLength(tags, result)
	}

	return result
}

func (v *TagValidator) checkTagCount(tags map[string]string, result *ComplianceResult) {
	if len(tags) > v.config.Global.TagCriteria.MaxTags {
		result.IsCompliant = false
		result.Violations = append(result.Violations, Violation{
			Type:    ViolationTypeExcessTags,
			Message: fmt.Sprintf("Number of tags (%d) exceeds maximum allowed (%d)", len(tags), v.config.Global.TagCriteria.MaxTags),
		})
	}
}

func (v *TagValidator) checkRequiredTags(tags map[string]string, result *ComplianceResult) {
	missingTags := []string{}
	for _, requiredTag := range v.config.Global.TagCriteria.RequiredTags {
		if _, exists := tags[requiredTag]; !exists {
			missingTags = append(missingTags, requiredTag)
		}
	}

	if len(missingTags) > 0 {
		result.IsCompliant = false
		result.Violations = append(result.Violations, Violation{
			Type:    ViolationTypeMissingTags,
			Message: fmt.Sprintf("Missing required tags: %v", missingTags),
		})
	}
}

func (v *TagValidator) checkProhibitedTags(tags map[string]string, result *ComplianceResult) {
	for tagKey := range tags {
		if contains(v.config.TagValidation.ProhibitedTags, tagKey) {
			result.IsCompliant = false
			result.Violations = append(result.Violations, Violation{
				Type:    ViolationTypeProhibitedTag,
				Message: fmt.Sprintf("Tag '%s' is prohibited", tagKey),
			})
		}
	}
}

func (v *TagValidator) validateKeyFormat(tags map[string]string, result *ComplianceResult) {
	for tagKey := range tags {
		for _, rule := range v.config.TagValidation.KeyFormatRules {
			matched, err := regexp.MatchString(rule.Pattern, tagKey)
			if err != nil || !matched {
				result.IsCompliant = false
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeInvalidKeyFormat,
					Message: fmt.Sprintf("Tag key '%s' does not match required format: %s", tagKey, rule.Message),
				})
			}
		}
	}
}

func (v *TagValidator) validateCaseRules(tags map[string]string, result *ComplianceResult) {
	for tagKey, caseRule := range v.config.TagValidation.CaseRules {
		tagValue, exists := tags[tagKey]
		if !exists {
			continue
		}

		var isValid bool
		switch caseRule.Case {
		case "lowercase":
			isValid = tagValue == strings.ToLower(tagValue)
		case "uppercase":
			isValid = tagValue == strings.ToUpper(tagValue)
		case "mixed":
			isValid = caseRule.Pattern == "" || regexp.MustCompile(caseRule.Pattern).MatchString(tagValue)
		default:
			isValid = true
		}

		if !isValid {
			result.IsCompliant = false
			result.Violations = append(result.Violations, Violation{
				Type:    ViolationTypeCaseViolation,
				Message: fmt.Sprintf("Tag %s violates case rule: %s", tagKey, caseRule.Message),
			})
		}
	}
}

func (v *TagValidator) validateValueLength(tags map[string]string, result *ComplianceResult) {
	for tagKey, value := range tags {
		if rule, exists := v.config.TagValidation.LengthRules[tagKey]; exists {
			if rule.MinLength != nil && len(value) < *rule.MinLength {
				result.IsCompliant = false
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeValueLength,
					Message: fmt.Sprintf("Tag '%s' value length (%d) is less than minimum required (%d)", tagKey, len(value), *rule.MinLength),
				})
			}
			if rule.MaxLength != nil && len(value) > *rule.MaxLength {
				result.IsCompliant = false
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeValueLength,
					Message: fmt.Sprintf("Tag '%s' value length (%d) exceeds maximum allowed (%d)", tagKey, len(value), *rule.MaxLength),
				})
			}
		}
	}
}

func (v *TagValidator) validateAllowedValues(tags map[string]string, result *ComplianceResult) {
	for tagKey, allowedValues := range v.config.TagValidation.AllowedValues {
		tagValue, exists := tags[tagKey]
		if !exists {
			continue
		}

		if !contains(allowedValues, tagValue) {
			result.IsCompliant = false
			result.Violations = append(result.Violations, Violation{
				Type:    ViolationTypeInvalidValue,
				Message: fmt.Sprintf("Tag %s has invalid value. Allowed values: %v", tagKey, allowedValues),
			})
		}
	}
}

func (v *TagValidator) validatePatternRules(tags map[string]string, result *ComplianceResult) {
	for tagKey, pattern := range v.config.TagValidation.PatternRules {
		tagValue, exists := tags[tagKey]
		if !exists {
			continue
		}

		matched, err := regexp.MatchString(pattern, tagValue)
		if err != nil || !matched {
			result.IsCompliant = false
			result.Violations = append(result.Violations, Violation{
				Type:    ViolationTypePatternViolation,
				Message: fmt.Sprintf("Tag %s does not match required pattern: %s", tagKey, pattern),
			})
		}
	}
}

// Helper function to check if a slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func (v *TagValidator) validateTagCase(tags map[string]string) []Violation {
	var violations []Violation

	for tagName, caseRule := range v.config.TagValidation.CaseRules {
		tagValue, exists := tags[tagName]
		if !exists {
			continue
		}

		var violation Violation
		switch caseRule.Case {
		case "lowercase":
			if tagValue != strings.ToLower(tagValue) {
				violation = Violation{
					Type:    "case_violation",
					Message: fmt.Sprintf("Tag %s violates case rule: %s", tagName, caseRule.Message),
				}
			}
		case "uppercase":
			if tagValue != strings.ToUpper(tagValue) {
				violation = Violation{
					Type:    "case_violation",
					Message: fmt.Sprintf("Tag %s violates case rule: %s", tagName, caseRule.Message),
				}
			}
		}

		if violation.Type != "" {
			violations = append(violations, violation)
		}
	}

	return violations
}
