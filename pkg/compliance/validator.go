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

	// New comprehensive validations
	v.validateKeyPrefixSuffix(tags, result)
	v.validateValueCharacters(tags, result)
	v.validateCaseSensitivity(tags, result)

	// Enhanced validation methods
	v.validateKeyFormat(tags, result)
	v.validateValueLength(tags, result)

	// Case rules validation
	hasCaseViolation := v.validateCaseRules(tags, result)
	if hasCaseViolation {
		return result
	}

	v.validateAllowedValues(tags, result)
	v.validatePatternRules(tags, result)

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
		for _, prohibitedTag := range v.config.TagValidation.ProhibitedTags {
			if strings.Contains(tagKey, prohibitedTag) {
				result.IsCompliant = false
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeProhibitedTag,
					Message: fmt.Sprintf("Tag '%s' contains prohibited prefix or substring: '%s'", tagKey, prohibitedTag),
				})
			}
		}
	}
}

func (v *TagValidator) validateKeyFormat(tags map[string]string, result *ComplianceResult) {
	keyFormatRules := v.config.TagValidation.KeyFormatRules

	if len(keyFormatRules) == 0 {
		return
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
			}
		}
	}
}

func (v *TagValidator) validateCaseRules(tags map[string]string, result *ComplianceResult) bool {
	caseRules := v.config.TagValidation.CaseRules

	if len(caseRules) == 0 {
		return false
	}

	for tagKey, caseRule := range caseRules {
		tagValue, exists := tags[tagKey]
		if !exists {
			continue
		}

		var isValid bool
		switch caseRule.Case {
		case configuration.CaseLowercase:
			isValid = tagValue == strings.ToLower(tagValue)
		case configuration.CaseUppercase:
			isValid = tagValue == strings.ToUpper(tagValue)
		case configuration.CaseMixed:
			// If pattern is specified, validate against it
			if caseRule.Pattern != "" {
				matched, err := regexp.MatchString(caseRule.Pattern, tagValue)
				if err != nil {
					isValid = false
				} else {
					isValid = matched
				}
			} else {
				// If no pattern, just allow mixed case
				isValid = true
			}
		default:
			isValid = true
		}

		if !isValid {
			result.IsCompliant = false
			result.Violations = append(result.Violations, Violation{
				Type:    ViolationTypeCaseViolation,
				Message: caseRule.Message,
				TagKey:  tagKey,
			})
			return true
		}
	}
	return false
}

func (v *TagValidator) validateValueLength(tags map[string]string, result *ComplianceResult) {
	lengthRules := v.config.TagValidation.LengthRules

	if len(lengthRules) == 0 {
		return
	}

	for tagKey, tagValue := range tags {
		if rule, exists := lengthRules[tagKey]; exists {
			if rule.MinLength != nil && len(tagValue) < *rule.MinLength {
				result.IsCompliant = false
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeValueLength,
					Message: rule.Message,
					TagKey:  tagKey,
				})
			}
			if rule.MaxLength != nil && len(tagValue) > *rule.MaxLength {
				result.IsCompliant = false
				result.Violations = append(result.Violations, Violation{
					Type:    ViolationTypeValueLength,
					Message: rule.Message,
					TagKey:  tagKey,
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
		fmt.Printf("Pattern validation - Tag: %s, Value: %s, Pattern: %s, Matched: %v, Error: %v\n",
			tagKey, tagValue, pattern, matched, err)

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
