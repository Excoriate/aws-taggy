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

	// Check minimum required tags
	if v.config.Global.TagCriteria.MinimumRequiredTags > 0 {
		v.checkRequiredTags(tags, result)
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

	return result
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

func (v *TagValidator) validateCaseRules(tags map[string]string, result *ComplianceResult) {
	for tagKey, caseRule := range v.config.TagValidation.CaseRules {
		tagValue, exists := tags[tagKey]
		if !exists {
			continue // Skip if tag doesn't exist
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

func (v *TagValidator) validateAllowedValues(tags map[string]string, result *ComplianceResult) {
	for tagKey, allowedValues := range v.config.TagValidation.AllowedValues {
		tagValue, exists := tags[tagKey]
		if !exists {
			continue // Skip if tag doesn't exist
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
			continue // Skip if tag doesn't exist
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
