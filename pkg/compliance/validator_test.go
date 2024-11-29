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
	}{
		{
			name: "Valid lowercase tags",
			tags: map[string]string{
				"environment": "development",
				"owner":       "team@company.com",
			},
			expectedResult:     true,
			expectedViolations: 0,
		},
		{
			name: "Invalid case for environment tag",
			tags: map[string]string{
				"Environment": "DEVELOPMENT",
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
		"Environment": "INVALID-ENV",
		"Owner":       "invalid-email",
	}

	result := validator.ValidateTags(tags)
	assert.False(t, result.IsCompliant)

	// Check for multiple violation types
	violationTypes := make(map[ViolationType]bool)
	for _, violation := range result.Violations {
		violationTypes[violation.Type] = true
		fmt.Printf("Violation: %+v\n", violation) // Debug print
	}

	assert.True(t, violationTypes[ViolationTypeCaseViolation], "Expected case violation")
	assert.True(t, violationTypes[ViolationTypePatternViolation], "Expected pattern violation")

	// Verify the number of violations
	assert.Len(t, result.Violations, 2, "Expected exactly 2 violations")
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
