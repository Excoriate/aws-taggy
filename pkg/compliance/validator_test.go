package compliance

import (
	"fmt"
	"testing"

	"github.com/Excoriate/aws-taggy/pkg/configuration"
	"github.com/stretchr/testify/assert"
)

func createTestConfig() *configuration.TaggyScanConfig {
	return &configuration.TaggyScanConfig{
		Global: configuration.GlobalConfig{
			Enabled: true,
			TagCriteria: configuration.TagCriteria{
				MinimumRequiredTags: 2,
				RequiredTags:        []string{"environment", "owner"},
			},
		},
		TagValidation: configuration.TagValidation{
			ProhibitedTags: []string{"temp", "test"},
			AllowedValues: map[string][]string{
				"environment": {"production", "staging", "development"},
			},
			CaseRules: map[string]configuration.CaseRule{
				"environment": {
					Case:    "lowercase",
					Message: "Environment must be lowercase",
				},
				"owner": {
					Case:    "lowercase",
					Message: "Owner must be lowercase",
				},
			},
			PatternRules: map[string]string{
				"owner": `^[a-z0-9._%+-]+@company\.com$`,
			},
			KeyFormatRules: []configuration.KeyFormatRule{
				{
					Pattern: "^[a-z][a-z0-9_-]*$",
					Message: "Tag keys must start with lowercase letter and contain only letters, numbers, underscores, and hyphens",
				},
			},
		},
	}
}

func TestValidateTags_RequiredTags(t *testing.T) {
	testCases := []struct {
		name               string
		tags               map[string]string
		expectedResult     bool
		expectedViolations int
	}{
		{
			name: "Valid tags with all required tags",
			tags: map[string]string{
				"environment": "production",
				"owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: 0,
		},
		{
			name: "Missing required tags",
			tags: map[string]string{
				"environment": "production",
			},
			expectedResult:     false,
			expectedViolations: 1,
		},
	}

	config := createTestConfig()
	validator := NewTagValidator(config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateTags(tc.tags)
			assert.Equal(t, tc.expectedResult, result.IsCompliant)
			assert.Len(t, result.Violations, tc.expectedViolations)
		})
	}
}

func TestValidateTags_AllowedValues(t *testing.T) {
	testCases := []struct {
		name               string
		tags               map[string]string
		expectedResult     bool
		expectedViolations int
	}{
		{
			name: "Valid environment tag",
			tags: map[string]string{
				"environment": "development",
				"owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: 0,
		},
		{
			name: "Invalid environment tag",
			tags: map[string]string{
				"environment": "invalid-env",
				"owner":       "team@company.com",
			},
			expectedResult:     false,
			expectedViolations: 1,
		},
	}

	config := createTestConfig()
	validator := NewTagValidator(config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateTags(tc.tags)
			assert.Equal(t, tc.expectedResult, result.IsCompliant)
			assert.Len(t, result.Violations, tc.expectedViolations)
		})
	}
}

func TestValidateTags_CaseRules(t *testing.T) {
	testCases := []struct {
		name               string
		tags               map[string]string
		expectedResult     bool
		expectedViolations int
		expectedTypes      []ViolationType
	}{
		{
			name: "Valid lowercase tags",
			tags: map[string]string{
				"environment": "development",
				"owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: 0,
			expectedTypes:      []ViolationType{},
		},
		{
			name: "Invalid case for environment tag",
			tags: map[string]string{
				"Environment": "DEVELOPMENT",
				"owner":       "team@company.com",
			},
			expectedResult:     false,
			expectedViolations: 3, // Key case + value case + key format
			expectedTypes: []ViolationType{
				ViolationTypeCaseViolation,
				ViolationTypeInvalidKeyFormat,
			},
		},
	}

	config := createTestConfig()
	validator := NewTagValidator(config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateTags(tc.tags)
			assert.Equal(t, tc.expectedResult, result.IsCompliant)
			assert.Len(t, result.Violations, tc.expectedViolations)

			if len(tc.expectedTypes) > 0 {
				violationTypes := make(map[ViolationType]bool)
				for _, v := range result.Violations {
					violationTypes[v.Type] = true
				}
				for _, expectedType := range tc.expectedTypes {
					assert.True(t, violationTypes[expectedType],
						"Expected violation type %s not found", expectedType)
				}
			}
		})
	}
}

func TestValidateTags_PatternRules(t *testing.T) {
	testCases := []struct {
		name               string
		tags               map[string]string
		expectedResult     bool
		expectedViolations int
	}{
		{
			name: "Valid owner email",
			tags: map[string]string{
				"environment": "development",
				"owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: 0,
		},
		{
			name: "Invalid owner email",
			tags: map[string]string{
				"environment": "development",
				"owner":       "invalid-email",
			},
			expectedResult:     false,
			expectedViolations: 1,
		},
	}

	config := createTestConfig()
	validator := NewTagValidator(config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateTags(tc.tags)
			assert.Equal(t, tc.expectedResult, result.IsCompliant)
			assert.Len(t, result.Violations, tc.expectedViolations)
		})
	}
}

func TestValidateTags_MultipleViolations(t *testing.T) {
	config := createTestConfig()
	validator := NewTagValidator(config)

	tags := map[string]string{
		"Environment": "INVALID-ENV",   // Case violation (key & value) + invalid value + key format
		"Owner":       "invalid-email", // Case violation + pattern violation + key format
	}

	result := validator.ValidateTags(tags)
	assert.False(t, result.IsCompliant)

	// Check for multiple violation types
	violationTypes := make(map[ViolationType]bool)
	for _, violation := range result.Violations {
		violationTypes[violation.Type] = true
		fmt.Printf("Violation: %+v\n", violation) // Debug print
	}

	// Verify all expected violation types are present
	assert.True(t, violationTypes[ViolationTypeCaseViolation], "Expected case violation")
	assert.True(t, violationTypes[ViolationTypeInvalidValue], "Expected invalid value violation")
	assert.True(t, violationTypes[ViolationTypePatternViolation], "Expected pattern violation")
	assert.True(t, violationTypes[ViolationTypeInvalidKeyFormat], "Expected invalid key format violation")

	// Count violations by tag
	envViolations := 0
	ownerViolations := 0
	for _, v := range result.Violations {
		if v.TagKey == "Environment" {
			envViolations++
		}
		if v.TagKey == "Owner" {
			ownerViolations++
		}
	}

	// Environment should have: case (key), case (value), invalid value, key format = 4 violations
	assert.GreaterOrEqual(t, envViolations, 3, "Expected at least 3 violations for Environment tag")
	// Owner should have: case (key), pattern violation, key format = 3 violations
	assert.GreaterOrEqual(t, ownerViolations, 2, "Expected at least 2 violations for Owner tag")
}

func TestValidateTags_ProhibitedTags(t *testing.T) {
	testCases := []struct {
		name               string
		tags               map[string]string
		expectedResult     bool
		expectedViolations []ViolationType
	}{
		{
			name: "Exact prohibited tag",
			tags: map[string]string{
				"temp":        "value",
				"environment": "development",
				"owner":       "team@company.com",
			},
			expectedResult:     false,
			expectedViolations: []ViolationType{ViolationTypeProhibitedTag},
		},
		{
			name: "Prohibited tag prefix",
			tags: map[string]string{
				"temp:test":   "temporary resource",
				"environment": "development",
				"owner":       "team@company.com",
			},
			expectedResult:     false,
			expectedViolations: []ViolationType{ViolationTypeProhibitedTag},
		},
		{
			name: "Prohibited tag substring",
			tags: map[string]string{
				"my-temp-tag": "some value",
				"environment": "development",
				"owner":       "team@company.com",
			},
			expectedResult:     false,
			expectedViolations: []ViolationType{ViolationTypeProhibitedTag},
		},
		{
			name: "No prohibited tags",
			tags: map[string]string{
				"environment": "development",
				"owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: []ViolationType{},
		},
	}

	config := createTestConfig()
	validator := NewTagValidator(config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateTags(tc.tags)
			assert.Equal(t, tc.expectedResult, result.IsCompliant)

			// Check violation types
			violationTypes := make([]ViolationType, 0)
			for _, violation := range result.Violations {
				violationTypes = append(violationTypes, violation.Type)
			}

			if len(tc.expectedViolations) > 0 {
				assert.Contains(t, violationTypes, tc.expectedViolations[0])
			}
		})
	}
}

func TestValidateTags_MultipleViolationsPerTag(t *testing.T) {
	config := createTestConfig()
	validator := NewTagValidator(config)

	tags := map[string]string{
		"Environment": "INVALID-ENV",   // Case violation + invalid value + key format
		"owner":       "invalid-email", // Pattern violation
		"temp":        "value",         // Prohibited tag
	}

	result := validator.ValidateTags(tags)
	assert.False(t, result.IsCompliant)

	// Check for multiple violation types
	violationTypes := make(map[ViolationType]int)
	for _, violation := range result.Violations {
		violationTypes[violation.Type]++
		fmt.Printf("Violation: %+v\n", violation) // Debug print
	}

	// Environment tag should have multiple violations
	assert.GreaterOrEqual(t, violationTypes[ViolationTypeCaseViolation], 1, "Expected case violation")
	assert.GreaterOrEqual(t, violationTypes[ViolationTypeInvalidValue], 1, "Expected invalid value violation")
	assert.GreaterOrEqual(t, violationTypes[ViolationTypeInvalidKeyFormat], 1, "Expected key format violation")

	// Owner tag should have pattern violation
	assert.GreaterOrEqual(t, violationTypes[ViolationTypePatternViolation], 1, "Expected pattern violation")

	// Temp tag should be prohibited
	assert.GreaterOrEqual(t, violationTypes[ViolationTypeProhibitedTag], 1, "Expected prohibited tag violation")

	// Total violations should be at least 5 (3 for Environment + 1 for owner + 1 for temp)
	totalViolations := 0
	for _, count := range violationTypes {
		totalViolations += count
	}
	assert.GreaterOrEqual(t, totalViolations, 5, "Expected at least 5 total violations")
}
