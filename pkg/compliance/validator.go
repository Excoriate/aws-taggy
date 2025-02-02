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

func (v *TagValidator) checkTagCount(tags map[string]string, result *ComplianceResult) bool {
	if v.config.Global.TagCriteria.MaxTags > 0 && len(tags) > v.config.Global.TagCriteria.MaxTags {
		result.IsCompliant = false
		result.Violations = append(result.Violations, Violation{
			Type:    ViolationTypeExcessTags,
			Message: fmt.Sprintf("Number of tags (%d) exceeds maximum allowed (%d)", len(tags), v.config.Global.TagCriteria.MaxTags),
		})
		return true
	}
	return false
}

func (v *TagValidator) checkAllowedValues(tags map[string]string, result *ComplianceResult) bool {
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
			return true
		}
	}
	return false
}

func (v *TagValidator) checkPatternRules(tags map[string]string, result *ComplianceResult) bool {
	for tagKey, pattern := range v.config.TagValidation.PatternRules {
		// Convert the tag key to lowercase for comparison
		lowercaseKey := strings.ToLower(tagKey)

		// Find the actual tag key in the map (case-insensitive)
		var actualKey string
		var tagValue string
		var exists bool

		for k, v := range tags {
			if strings.ToLower(k) == lowercaseKey {
				actualKey = k
				tagValue = v
				exists = true
				break
			}
		}

		if !exists {
			continue
		}

		matched, err := regexp.MatchString(pattern, tagValue)
		if err != nil || !matched {
			result.IsCompliant = false
			result.Violations = append(result.Violations, Violation{
				Type:    ViolationTypePatternViolation,
				Message: fmt.Sprintf("Tag %s does not match required pattern: %s", actualKey, pattern),
				TagKey:  actualKey,
			})
			return true
		}
	}
	return false
}

func (v *TagValidator) checkKeyFormatRules(tags map[string]string, result *ComplianceResult) bool {
	keyFormatRules := v.config.TagValidation.KeyFormatRules

	if len(keyFormatRules) == 0 {
		return false
	}

	for tagKey := range tags {
		for _, rule := range keyFormatRules {
			matched, err := regexp.MatchString(rule.Pattern, tagKey)
			if err != nil || !matched {
				result.IsCompliant = false
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeInvalidKeyFormat,
					Message: rule.Message,
					TagKey:  tagKey,
				})
				return true
			}
		}
	}
	return false
}

func (v *TagValidator) validateCaseRules(tags map[string]string, result *ComplianceResult) {
	for key, value := range tags {
		for ruleKey, caseRule := range v.config.TagValidation.CaseRules {
			if strings.EqualFold(key, ruleKey) {
				// Check key case
				if key != ruleKey {
					result.Violations = append(result.Violations, Violation{
						Type:    ViolationTypeCaseViolation,
						Message: fmt.Sprintf("Tag key '%s' must match case '%s'", key, ruleKey),
						TagKey:  key,
					})
				}

				// Check value case
				switch caseRule.Case {
				case configuration.CaseLowercase:
					if value != strings.ToLower(value) {
						result.Violations = append(result.Violations, Violation{
							Type:    ViolationTypeCaseViolation,
							Message: fmt.Sprintf("Tag value for '%s' must be lowercase", key),
							TagKey:  key,
						})
					}
				case configuration.CaseUppercase:
					if value != strings.ToUpper(value) {
						result.Violations = append(result.Violations, Violation{
							Type:    ViolationTypeCaseViolation,
							Message: fmt.Sprintf("Tag value for '%s' must be uppercase", key),
							TagKey:  key,
						})
					}
				}
			}
		}
	}
}

func (v *TagValidator) validatePatternRules(tags map[string]string, result *ComplianceResult) {
	for key, value := range tags {
		for ruleKey, pattern := range v.config.TagValidation.PatternRules {
			if strings.EqualFold(key, ruleKey) {
				matched, err := regexp.MatchString(pattern, value)
				if err != nil {
					// Log the error but continue validation
					log.Printf("Error matching pattern for tag %s: %v", key, err)
					continue
				}
				if !matched {
					result.Violations = append(result.Violations, Violation{
						Type:    ViolationTypePatternViolation,
						Message: fmt.Sprintf("Tag value for '%s' does not match required pattern", key),
						TagKey:  key,
					})
				}
			}
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

func (v *TagValidator) validateKeyPrefixSuffix(tags map[string]string, result *ComplianceResult) {
	keyValidation := v.config.TagValidation.KeyValidation

	// Check if KeyValidation struct is empty
	if len(keyValidation.AllowedPrefixes) == 0 &&
		len(keyValidation.AllowedSuffixes) == 0 &&
		keyValidation.MaxLength == 0 {
		return
	}

	for tagKey := range tags {
		// Prefix validation
		if len(keyValidation.AllowedPrefixes) > 0 {
			prefixValid := false
			for _, prefix := range keyValidation.AllowedPrefixes {
				if strings.HasPrefix(tagKey, prefix) {
					prefixValid = true
					break
				}
			}
			if !prefixValid {
				result.IsCompliant = false
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeInvalidKeyFormat,
					Message: fmt.Sprintf("Tag key '%s' does not have an allowed prefix", tagKey),
				})
			}
		}

		// Suffix validation
		if len(keyValidation.AllowedSuffixes) > 0 {
			suffixValid := false
			for _, suffix := range keyValidation.AllowedSuffixes {
				if strings.HasSuffix(tagKey, suffix) {
					suffixValid = true
					break
				}
			}
			if !suffixValid {
				result.IsCompliant = false
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeInvalidKeyFormat,
					Message: fmt.Sprintf("Tag key '%s' does not have an allowed suffix", tagKey),
				})
			}
		}

		// Max length validation
		if keyValidation.MaxLength > 0 && len(tagKey) > keyValidation.MaxLength {
			result.IsCompliant = false
			result.Violations = append(result.Violations, Violation{
				Type:    ViolationTypeInvalidKeyFormat,
				Message: fmt.Sprintf("Tag key '%s' exceeds maximum length of %d", tagKey, keyValidation.MaxLength),
			})
		}
	}
}

func (v *TagValidator) validateValueCharacters(tags map[string]string, result *ComplianceResult) {
	valueValidation := v.config.TagValidation.ValueValidation

	// Check if ValueValidation struct is empty
	if valueValidation.AllowedCharacters == "" &&
		len(valueValidation.DisallowedValues) == 0 {
		return
	}

	// If allowed characters are specified, create a regex
	var allowedCharsRegex *regexp.Regexp
	if valueValidation.AllowedCharacters != "" {
		allowedCharsRegex = regexp.MustCompile(fmt.Sprintf("^[%s]+$", valueValidation.AllowedCharacters))
	}

	for tagKey, tagValue := range tags {
		// Validate allowed characters if regex is defined
		if allowedCharsRegex != nil && !allowedCharsRegex.MatchString(tagValue) {
			result.IsCompliant = false
			result.Violations = append(result.Violations, Violation{
				Type:    ViolationTypeInvalidValue,
				Message: fmt.Sprintf("Tag '%s' contains disallowed characters. Only %s are allowed", tagKey, valueValidation.AllowedCharacters),
			})
		}

		// Disallowed values check
		if len(valueValidation.DisallowedValues) > 0 {
			for _, disallowedValue := range valueValidation.DisallowedValues {
				if tagValue == disallowedValue {
					result.IsCompliant = false
					result.Violations = append(result.Violations, Violation{
						Type:    ViolationTypeInvalidValue,
						Message: fmt.Sprintf("Tag '%s' contains disallowed value: %s", tagKey, disallowedValue),
					})
				}
			}
		}
	}
}

func (v *TagValidator) validateCaseSensitivity(tags map[string]string, result *ComplianceResult) {
	caseSensitivity := v.config.TagValidation.CaseSensitivity

	// Check if CaseSensitivity map is empty
	if len(caseSensitivity) == 0 {
		return
	}

	for tagKey, tagValue := range tags {
		if caseSensitivityRule, exists := caseSensitivity[tagKey]; exists {
			switch caseSensitivityRule.Mode {
			case "strict":
				// Exact case matching
				if tagValue != v.preserveOriginalCase(tagValue) {
					result.IsCompliant = false
					result.Violations = append(result.Violations, Violation{
						Type:    ViolationTypeCaseViolation,
						Message: fmt.Sprintf("Tag '%s' must maintain exact original case", tagKey),
					})
				}
			case "relaxed":
				// More lenient case validation
				if strings.ToLower(tagValue) != strings.ToLower(v.preserveOriginalCase(tagValue)) {
					result.IsCompliant = false
					result.Violations = append(result.Violations, Violation{
						Type:    ViolationTypeCaseViolation,
						Message: fmt.Sprintf("Tag '%s' has case inconsistency", tagKey),
					})
				}
			}
		}
	}
}

func (v *TagValidator) preserveOriginalCase(s string) string {
	// Placeholder method to preserve original case
	// In a real implementation, this would track the original input
	return s
}
