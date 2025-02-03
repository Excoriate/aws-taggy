package compliance

import (
	"fmt"
	"log"
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
		Violations:   make([]Violation, 0),
		ResourceTags: tags,
	}

	// Check tag count first
	if v.config.Global.TagCriteria.MaxTags > 0 && len(tags) > v.config.Global.TagCriteria.MaxTags {
		result.Violations = append(result.Violations, Violation{
			Type:    ViolationTypeExcessTags,
			Message: fmt.Sprintf("Number of tags (%d) exceeds maximum allowed (%d)", len(tags), v.config.Global.TagCriteria.MaxTags),
		})
		result.IsCompliant = false
	}

	// Check required tags
	missingTags := v.checkRequiredTags(tags)
	if len(missingTags) > 0 {
		result.Violations = append(result.Violations, Violation{
			Type:    ViolationTypeMissingTags,
			Message: fmt.Sprintf("Missing required tags: %v", missingTags),
		})
		result.IsCompliant = false
	}

	// Check prohibited tags
	for key := range tags {
		if v.isProhibitedTag(key) {
			result.Violations = append(result.Violations, Violation{
				Type:    ViolationTypeProhibitedTag,
				Message: fmt.Sprintf("Tag '%s' is prohibited", key),
				TagKey:  key,
			})
			result.IsCompliant = false
		}
	}

	// Validate case rules and key format for all tags
	for key, value := range tags {
		// Check key format rules
		for _, rule := range v.config.TagValidation.KeyFormatRules {
			matched, err := regexp.MatchString(rule.Pattern, key)
			if err != nil {
				log.Printf("Error matching key format pattern for tag %s: %v", key, err)
				continue
			}
			if !matched {
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeInvalidKeyFormat,
					Message: fmt.Sprintf("Tag key '%s': %s", key, rule.Message),
					TagKey:  key,
				})
				result.IsCompliant = false
			}
		}

		// Check case rules
		for ruleKey, caseRule := range v.config.TagValidation.CaseRules {
			if strings.EqualFold(key, ruleKey) {
				// Check key case
				if key != strings.ToLower(ruleKey) {
					result.Violations = append(result.Violations, Violation{
						Type:    ViolationTypeCaseViolation,
						Message: fmt.Sprintf("Tag key '%s' must match case '%s'", key, strings.ToLower(ruleKey)),
						TagKey:  key,
					})
					result.IsCompliant = false
				}

				// Check value case
				switch caseRule.Case {
				case "lowercase":
					if value != strings.ToLower(value) {
						result.Violations = append(result.Violations, Violation{
							Type:    ViolationTypeCaseViolation,
							Message: fmt.Sprintf("Tag value for '%s' must be lowercase", key),
							TagKey:  key,
						})
						result.IsCompliant = false
					}
				case "uppercase":
					if value != strings.ToUpper(value) {
						result.Violations = append(result.Violations, Violation{
							Type:    ViolationTypeCaseViolation,
							Message: fmt.Sprintf("Tag value for '%s' must be uppercase", key),
							TagKey:  key,
						})
						result.IsCompliant = false
					}
				}
			}
		}

		// Check pattern rules
		for ruleKey, pattern := range v.config.TagValidation.PatternRules {
			if strings.EqualFold(key, ruleKey) {
				matched, err := regexp.MatchString(pattern, value)
				if err != nil {
					log.Printf("Error matching pattern for tag %s: %v", key, err)
					continue
				}
				if !matched {
					result.Violations = append(result.Violations, Violation{
						Type:    ViolationTypePatternViolation,
						Message: fmt.Sprintf("Tag value for '%s' does not match required pattern", key),
						TagKey:  key,
					})
					result.IsCompliant = false
				}
			}
		}

		// Check allowed values
		if allowedValues, exists := v.config.TagValidation.AllowedValues[strings.ToLower(key)]; exists {
			valueAllowed := false
			for _, allowedValue := range allowedValues {
				if strings.EqualFold(value, allowedValue) {
					valueAllowed = true
					break
				}
			}
			if !valueAllowed {
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeInvalidValue,
					Message: fmt.Sprintf("Tag value for '%s' must be one of: %v", key, allowedValues),
					TagKey:  key,
				})
				result.IsCompliant = false
			}
		}
	}

	return result
}

func (v *TagValidator) checkRequiredTags(tags map[string]string) []string {
	var missingTags []string
	for _, requiredTag := range v.config.Global.TagCriteria.RequiredTags {
		found := false
		for tagKey := range tags {
			if strings.EqualFold(tagKey, requiredTag) {
				found = true
				break
			}
		}
		if !found {
			missingTags = append(missingTags, requiredTag)
		}
	}
	return missingTags
}

func (v *TagValidator) isProhibitedTag(tagKey string) bool {
	for _, prohibitedTag := range v.config.TagValidation.ProhibitedTags {
		if strings.Contains(strings.ToLower(tagKey), strings.ToLower(prohibitedTag)) {
			return true
		}
	}
	return false
}
